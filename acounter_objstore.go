package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	pcounter "github.com/synerex/proto_pcounter"
	storage "github.com/synerex/proto_storage"
	api "github.com/synerex/synerex_api"
	pbase "github.com/synerex/synerex_proto"

	sxutil "github.com/synerex/synerex_sxutil"
	//sxutil "local.packages/synerex_sxutil"

	"log"
	"sync"
)

// datastore provider provides Datastore Service.

var (
	nodesrv         = flag.String("nodesrv", "127.0.0.1:9990", "Node ID Server")
	local           = flag.String("local", "", "Local Synerex Server")
	mu              sync.Mutex
	version         = "0.01"
	baseDir         = "store"
	dataDir         string
	pcMu            *sync.Mutex = nil
	pcLoop          *bool       = nil
	ssMu            *sync.Mutex = nil
	ssLoop          *bool       = nil
	sxServerAddress string
	currentNid      uint64                  = 0 // NotifyDemand message ID
	mbusID          uint64                  = 0 // storage MBus ID
	storageID       uint64                  = 0 // storageID
	pfClient        *sxutil.SXServiceClient = nil
	stClient        *sxutil.SXServiceClient = nil
	acblocks        map[string]*ACBlock     = map[string]*ACBlock{}
	iacblocks       map[string]*IACBlock    = map[string]*IACBlock{}
	bucketName                              = flag.String("bucket", "centrair", "Bucket Name")
	holdPeriod                              = flag.Int64("holdPeriod", 30, "ACounter Data Hold Time")
)

const layout = "2006-01-02T15:04:05.999999Z"

func init() {
}

func objStore(bc string, ob string, dt string) {

	log.Printf("Store %s, %s, %s", bc, ob, dt)
	//  we need to send data into mbusID.
	record := storage.Record{
		BucketName: bc,
		ObjectName: ob,
		Record:     []byte(dt),
		Option:     []byte("raw"),
	}
	out, err := proto.Marshal(&record)
	if err == nil {
		cont := &api.Content{Entity: out}
		smo := sxutil.SupplyOpts{
			Name:  "Record", // command
			Cdata: cont,
		}
		stClient.NotifySupply(&smo)
	}

}

// saveRecursive : save to objstorage recursive
func saveRecursive(client *sxutil.SXServiceClient) {
	// ch := make(chan error)
	for {
		time.Sleep(time.Second * time.Duration(60))
		currentTime := time.Now().Unix() + 9*3600
		log.Printf("\nCurrent: %d", currentTime)
		for name, acblock := range acblocks {
			if acblock.BaseDate+*holdPeriod < currentTime {
				data, err := json.Marshal(acblock.ACounters)

				if err == nil {
					objStore(*bucketName, name, string(data)+"\n")
					delete(acblocks, name)
				} else {
					log.Printf("Error!!: %+v\n", err)
				}
			}
		}
		for name, iacblock := range iacblocks {
			if iacblock.BaseDate+*holdPeriod < currentTime {
				aclines := []string{}
				for _, ac := range iacblock.ACounter {
					st, _ := time.Parse(layout, ptypes.TimestampString(ac.Ts))
					aclines = append(aclines, fmt.Sprintf("%s,%d,%s,%d", st.Format(layout), ac.AreaId, ac.AreaName, ac.Count))
				}
				objStore(*bucketName, name, strings.Join(aclines, "\n")+"\n")
				delete(iacblocks, name)
			}
		}
	}
}

// called for each agent data.
func supplyACounterCallback(clt *sxutil.SXServiceClient, sp *api.Supply) {

	pc := &pcounter.ACounters{}

	err := proto.Unmarshal(sp.Cdata.Entity, pc)
	if err == nil { // get ACounter
		tsd, _ := ptypes.Timestamp(pc.Ts)
		tsd = tsd.Add(9 * time.Hour)
		log.Printf("%v", pc)

		// how to define Bucket:

		// we use IP address for sensor_id
		//		objectName := "year/month/date/hour/min"
		objectName := fmt.Sprintf("ACOUNTER/ACOUNTER/%4d/%02d/%02d/%02d/%02d", tsd.Year(), tsd.Month(), tsd.Day(), tsd.Hour(), tsd.Minute())

		if acblock, exists := acblocks[objectName]; exists {
			acblock.ACounters = append(acblock.ACounters, pc)
		} else {
			acblocks[objectName] = &ACBlock{
				BaseDate:  tsd.Unix(),
				ACounters: []*pcounter.ACounters{pc},
			}
		}

		for _, ac := range pc.Acs {
			objectName := fmt.Sprintf("AREA/%s/%4d/%02d/%02d/%02d/%02d", ac.AreaName, tsd.Year(), tsd.Month(), tsd.Day(), tsd.Hour(), tsd.Minute())

			if iacblock, exists := iacblocks[objectName]; exists {
				iacblock.ACounter = append(iacblock.ACounter, ac)
			} else {
				iacblocks[objectName] = &IACBlock{
					BaseDate: tsd.Unix(),
					ACounter: []*pcounter.ACounter{ac},
				}
			}
		}
	}
}

func main() {
	flag.Parse()
	go sxutil.HandleSigInt()
	sxutil.RegisterDeferFunction(sxutil.UnRegisterNode)
	log.Printf("ACounter-ObjStorage(%s) built %s sha1 %s", sxutil.GitVer, sxutil.BuildTime, sxutil.Sha1Ver)

	channelTypes := []uint32{pbase.AREA_COUNTER_SVC, pbase.STORAGE_SERVICE}

	var rerr error
	sxServerAddress, rerr = sxutil.RegisterNode(*nodesrv, "ACounterObjStorage", channelTypes, nil)

	if rerr != nil {
		log.Fatal("Can't register node:", rerr)
	}
	if *local != "" { // quick hack for AWS local network
		sxServerAddress = *local
	}
	log.Printf("Connecting SynerexServer at [%s]", sxServerAddress)

	wg := sync.WaitGroup{} // for syncing other goroutines

	client := sxutil.GrpcConnectServer(sxServerAddress)

	if client == nil {
		log.Fatal("Can't connect Synerex Server")
	}

	stClient = sxutil.NewSXServiceClient(client, pbase.STORAGE_SERVICE, "{Client:ACObjStore}")
	pfClient = sxutil.NewSXServiceClient(client, pbase.AREA_COUNTER_SVC, "{Client:ACObjStore}")

	log.Print("Subscribe ACounter Supply")
	pcMu, pcLoop = sxutil.SimpleSubscribeSupply(pfClient, supplyACounterCallback)
	wg.Add(1)

	go saveRecursive(pfClient)

	wg.Wait()

}
