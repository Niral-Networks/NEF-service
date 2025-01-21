package datacollection

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/free5gc/openapi"
	"github.com/free5gc/openapi/Nnrf_NFDiscovery"
	"github.com/free5gc/openapi/models"
	"github.com/geekaamit/NEF-service/consumer"
	nef_context "github.com/geekaamit/NEF-service/context"
	"github.com/geekaamit/NEF-service/factory"
	"github.com/geekaamit/NEF-service/model"
	"github.com/geekaamit/NEF-service/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var Nfid string
var subsId string
var m1 = make(map[string]string)
var m2 = make(map[string]string)

func HTTPAmfRegistrationAccept(c *gin.Context) {
	var registrationAccept model.RegistrationAccept
	requestBody, err := c.GetRawData()
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadGateway)
		c.Writer.Write([]byte("Internal Error"))
		return
	}

	err = openapi.Deserialize(&registrationAccept, requestBody, "application/json")
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadGateway)
		c.Writer.Write([]byte("Json Parser Error"))
		return
	}

	registrationAccept.Date = time.Now()
	/* registrar na base */
	util.AddRegistrationAccept(&registrationAccept)
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Write([]byte("Ok"))
}
func SendHTTPAmfRegistrationRequest(c *gin.Context) {
	self := nef_context.NEF_Self()
	searchOpt := Nnrf_NFDiscovery.SearchNFInstancesParamOpts{}

	var createEventSubscription models.AmfCreateEventSubscription
	//var amfeventnotifylist model.AmfEventnotifylist
	//var eventsubdata model.NefEventSubscriptionData
	requestBody, err := c.GetRawData()
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadGateway)
		c.Writer.Write([]byte("Internal Error"))
		return
	}

	err = openapi.Deserialize(&createEventSubscription, requestBody, "application/json")
	//createEventSubscription.Subscription.EventNotifyUri = "http://127.0.0.55:29599/datacollection/amf-contexts/registration-accept"
	if len(m1) == 0 && len(m2) == 0 {
		// self.SubscriptionDataSubscriptionIDGenerator = 1
		// id := self.SubscriptionDataSubscriptionIDGenerator
		// newSubscriptionID := strconv.Itoa(int(id))
		subsId = uuid.New().String()
		//subsId = newSubscriptionID
		createEventSubscription.Subscription.SubscriptionCorrelationId = subsId
		createEventSubscription.Subscription.NfId = Nfid
		//url := createEventSubscription.Subscription.EventNotifyUri
		m1[subsId] = createEventSubscription.Subscription.EventNotifyUri
		m2[subsId] = createEventSubscription.Subscription.NfId
	} else {
		for subid, nfid := range m2 {
			if nfid == createEventSubscription.Subscription.NfId {
				delete(m1, subid)
				delete(m2, subid)
			}
		}
		subsId = uuid.New().String()
		createEventSubscription.Subscription.SubscriptionCorrelationId = subsId
		createEventSubscription.Subscription.NfId = Nfid
		//url := createEventSubscription.Subscription.EventNotifyUri
		m1[subsId] = createEventSubscription.Subscription.EventNotifyUri
		m2[subsId] = createEventSubscription.Subscription.NfId
	}
	//err = openapi.Deserialize(&amfeventnotifylist, requestBody, "application/json")
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadGateway)
		c.Writer.Write([]byte("Json Parser Error"))
		return
	}
	//amfeventnotifylist.Subscription.EventNotifyUri = "http://127.0.0.55:29599/datacollection/amf-contexts/registration-accept"
	apiPrefix := fmt.Sprintf("%s://%s:%d/%s", self.UriScheme, self.RegisterIPv4, self.SBIPort, "datacollection/amf-contexts/registration-accept")
	createEventSubscription.Subscription.EventNotifyUri = apiPrefix
	requestBody, err = openapi.Serialize(&createEventSubscription, "application/json")
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadGateway)
		c.Writer.Write([]byte("Json Parser Error"))
		return
	}
	// recupera todas as AMFs registradas na NRF
	resp, err := consumer.SendSearchNFInstances(self.NrfUri, models.NfType_AMF, models.NfType_NEF, searchOpt)
	//resp, err := consumer.SendSearchNFInstances(self.NrfUri, models.NfType_NEF, models.NfType_NWDAF, searchOpt)
	if err != nil {
		fmt.Println(err)
	}

	//para cada uma das AMF's registrar no core realiza o subscriber de coleta
	for _, nfProfile := range resp.NfInstances {

		/* localiza a URL do end-point de subscriber com status de REGISTRADO */
		amfUri, endpoint, apiversion := util.SearchNFServiceUri(nfProfile, models.ServiceName_NAMF_EVTS, models.NfServiceStatus_REGISTERED)
		//amfUri, endpoint, apiversion := util.SearchNFServiceUri(nfProfile, models.ServiceName_NNEF_EVENTSSUBSCRIPTION, models.NfServiceStatus_REGISTERED)

		fmt.Println(endpoint)
		fmt.Println(apiversion)

		var buffer bytes.Buffer

		buffer.WriteString(amfUri)
		buffer.WriteString("/")
		buffer.WriteString(endpoint)
		buffer.WriteString("/")
		buffer.WriteString(apiversion)
		buffer.WriteString("/")
		buffer.WriteString("subscriptions")

		url := buffer.String()

		/*
		 * 1 º os possiveis tipos de eventos p/ AMF estão em AmfEventType
		 */

		// 	var jsonStr = []byte(jsonData)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("X-Custom-Header", "myvalue")
		req.Header.Set("Content-Type", "application/json")
		serverScheme := factory.NefConfig.Configuration.Sbi.Scheme
		var state bool = strings.Contains(serverScheme, "https")
		var certFile = factory.NefConfig.Configuration.Sbi.Tls.Pem
		var keyFile = factory.NefConfig.Configuration.Sbi.Tls.Key
		client := util.GetHttpConnection(state, certFile, keyFile)

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body) // response body is []byte
		fmt.Println(string(body))
	}

}

func SendHTTPAmfRegistrationResponseAMF(c *gin.Context) {
	var amfeventnotifylist model.AmfEventnotifylist
	requestBody, err := c.GetRawData()
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadGateway)
		c.Writer.Write([]byte("Internal Error"))
		return
	}
	err = openapi.Deserialize(&amfeventnotifylist, requestBody, "application/json")
	if err != nil {
		log.Fatal(err)
	}
	for id, url := range m1 {
		if id == amfeventnotifylist.Subscription.SubscriptionCorrelationId {
			amfeventnotifylist.Subscription.EventNotifyUri = url
		}
	}
	URL := amfeventnotifylist.Subscription.EventNotifyUri
	requestBody, err = openapi.Serialize(&amfeventnotifylist, "application/json")
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadGateway)
		c.Writer.Write([]byte("Json Parser Error"))
		return
	}
	// req, err := http.NewRequest("POST", URL, bytes.NewBuffer(requestBody))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// serverScheme := factory.NefConfig.Configuration.Sbi.Scheme
	// var state bool = strings.Contains(serverScheme, "https")
	// var certFile = factory.NefConfig.Configuration.Sbi.Tls.Pem
	// var keyFile = factory.NefConfig.Configuration.Sbi.Tls.Key
	// client := util.GetHttpConnection(state, certFile, keyFile)
	// resp, err := client.Do(req)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer resp.Body.Close()

	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	client := &http.Client{
		Transport: transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()
}

func SendAmfEventUnsubscribeRequest(c *gin.Context) {
	var createEventSubscription models.AmfCreateEventSubscription
	searchOpt := Nnrf_NFDiscovery.SearchNFInstancesParamOpts{}
	self := nef_context.NEF_Self()
	requestBody, err := c.GetRawData()
	if err != nil {
		c.Writer.WriteHeader(http.StatusBadGateway)
		c.Writer.Write([]byte("Internal Error"))
		return
	}

	err = openapi.Deserialize(&createEventSubscription, requestBody, "application/json")
	if err != nil {
		log.Fatal(err)
	}
	subscriptionId := createEventSubscription.Subscription.SubscriptionCorrelationId
	for id := range m1 {
		if id == createEventSubscription.Subscription.SubscriptionCorrelationId {
			resp, err := consumer.SendSearchNFInstances(self.NrfUri, models.NfType_AMF, models.NfType_NEF, searchOpt)
			//resp, err := consumer.SendSearchNFInstances(self.NrfUri, models.NfType_NEF, models.NfType_NWDAF, searchOpt)
			if err != nil {
				fmt.Println(err)
			}

			//para cada uma das AMF's registrar no core realiza o subscriber de coleta
			for _, nfProfile := range resp.NfInstances {

				/* localiza a URL do end-point de subscriber com status de REGISTRADO */
				amfUri, endpoint, apiversion := util.SearchNFServiceUri(nfProfile, models.ServiceName_NAMF_EVTS, models.NfServiceStatus_REGISTERED)
				//amfUri, endpoint, apiversion := util.SearchNFServiceUri(nfProfile, models.ServiceName_NNEF_EVENTSSUBSCRIPTION, models.NfServiceStatus_REGISTERED)

				fmt.Println(endpoint)
				fmt.Println(apiversion)

				var buffer bytes.Buffer

				buffer.WriteString(amfUri)
				buffer.WriteString("/")
				buffer.WriteString(endpoint)
				buffer.WriteString("/")
				buffer.WriteString(apiversion)
				buffer.WriteString("/")
				buffer.WriteString("unsubscribe")
				buffer.WriteString("/")
				buffer.WriteString(subscriptionId)

				url := buffer.String()

				/*
				 * 1 º os possiveis tipos de eventos p/ AMF estão em AmfEventType
				 */

				// 	var jsonStr = []byte(jsonData)
				req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(requestBody))
				if err != nil {
					log.Fatal(err)
				}
				req.Header.Set("X-Custom-Header", "myvalue")
				req.Header.Set("Content-Type", "application/json")

				serverScheme := factory.NefConfig.Configuration.Sbi.Scheme
				var state bool = strings.Contains(serverScheme, "https")
				var certFile = factory.NefConfig.Configuration.Sbi.Tls.Pem
				var keyFile = factory.NefConfig.Configuration.Sbi.Tls.Key
				client := util.GetHttpConnection(state, certFile, keyFile)

				resp, err := client.Do(req)
				if err != nil {
					log.Fatal(err)
				}
				defer resp.Body.Close()
			}
		}
	}
}

func GetNfId(self *nef_context.NEFContext) string {
	Nfid = self.NfId

	return Nfid
}

func NWDAFEventUnsubscribe() {

}
