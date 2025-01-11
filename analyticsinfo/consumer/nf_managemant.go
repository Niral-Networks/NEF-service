package consumer

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	nef_context "github.com/NEF-service/context"
	"github.com/NEF-service/factory"
	"github.com/free5gc/openapi/Nnrf_NFManagement"
	"github.com/free5gc/openapi/models"
)

func BuildNFInstance(context *nef_context.NEFContext) models.NfProfile {
	var profile models.NfProfile
	config := factory.NefConfig
	profile.NfInstanceId = context.NfId
	profile.NfType = models.NfType_NEF
	profile.NfStatus = models.NfStatus_REGISTERED
	profile.Ipv4Addresses = []string{context.RegisterIPv4}
	profile.AllowedNfTypes = []models.NfType{models.NfType_AMF, models.NfType_SMF, models.NfType_NWDAF}
	version := config.Info.Version
	tmpVersion := strings.Split(version, ".")
	versionUri := "v" + tmpVersion[0]
	apiPrefix := fmt.Sprintf("%s://%s:%d", context.UriScheme, context.RegisterIPv4, context.SBIPort)
	services := []models.NfService{ //TODO: Outras funções usam um "for" para preencher os serviços.
		{
			ServiceInstanceId: "nefdatarepository", //TODO: Renomear para o ID correto. E excluir o código do serviço de exemplo: ServiceName_NNEF_DR
			ServiceName:       "nnef-dr",           //TODO: Renomear para o serviço correto! ServiceName_NNEF_ANALYTICSINFO
			Versions: &[]models.NfServiceVersion{
				{
					ApiFullVersion:  version,
					ApiVersionInUri: versionUri,
				},
			},
			Scheme:          models.UriScheme(context.UriScheme),
			NfServiceStatus: models.NfServiceStatus_REGISTERED,
			ApiPrefix:       apiPrefix,
			IpEndPoints: &[]models.IpEndPoint{
				{
					Ipv4Address: context.RegisterIPv4,
					Transport:   models.TransportProtocol_TCP,
					Port:        int32(context.SBIPort),
				},
			},
		},
		{
			ServiceInstanceId: "analyticsinfo",
			ServiceName:       models.ServiceName_NNEF_ANALYTICSINFO,
			Versions: &[]models.NfServiceVersion{
				{
					ApiFullVersion:  version,
					ApiVersionInUri: versionUri,
				},
			},
			Scheme:          models.UriScheme(context.UriScheme),
			NfServiceStatus: models.NfServiceStatus_REGISTERED,
			ApiPrefix:       apiPrefix,
			IpEndPoints: &[]models.IpEndPoint{
				{
					Ipv4Address: context.RegisterIPv4,
					Transport:   models.TransportProtocol_TCP,
					Port:        int32(context.SBIPort),
				},
			},
		},
		{
			ServiceInstanceId: "eventssubscription",
			ServiceName:       models.ServiceName_NNEF_EVENTSSUBSCRIPTION,
			Versions: &[]models.NfServiceVersion{
				{
					ApiFullVersion:  version,
					ApiVersionInUri: versionUri,
				},
			},
			Scheme:          models.UriScheme(context.UriScheme),
			NfServiceStatus: models.NfServiceStatus_REGISTERED,
			ApiPrefix:       apiPrefix,
			IpEndPoints: &[]models.IpEndPoint{
				{
					Ipv4Address: context.RegisterIPv4,
					Transport:   models.TransportProtocol_TCP,
					Port:        int32(context.SBIPort),
				},
			},
		},
	}
	profile.NfServices = &services
	// TODO: finish the nef Info
	/*profile.NefInfo = &models.NefInfo{
		SupportedDataSets: []models.DataSetId{
			// models.DataSetId_APPLICATION,
			// models.DataSetId_EXPOSURE,
			// models.DataSetId_POLICY,
			models.DataSetId_SUBSCRIPTION,
		},
	}*/
	return profile
}

func SendRegisterNFInstance(nrfUri, nfInstanceId string, profile models.NfProfile) (string, string, error) {

	// Set client and set url
	configuration := Nnrf_NFManagement.NewConfiguration()
	configuration.SetBasePath(nrfUri)
	client := Nnrf_NFManagement.NewAPIClient(configuration)
	var resouceNrfUri string
	var retrieveNfInstanceId string
	var certFile = factory.NefConfig.Configuration.Sbi.Tls.Pem
	var keyFile = factory.NefConfig.Configuration.Sbi.Tls.Key
	for {
		_, res, err := client.NFInstanceIDDocumentApi.RegisterNFInstance(context.TODO(), nfInstanceId, profile, certFile, keyFile)
		if err != nil || res == nil {
			//TODO : add log
			fmt.Println(fmt.Errorf("NEF register to NRF Error[%s]", err.Error()))
			time.Sleep(2 * time.Second)
			continue
		}
		status := res.StatusCode
		if status == http.StatusOK {
			// NFUpdate
			return resouceNrfUri, retrieveNfInstanceId, err
		} else if status == http.StatusCreated {
			// NFRegister
			resourceUri := res.Header.Get("Location")
			resouceNrfUri = resourceUri[:strings.Index(resourceUri, "/nnrf-nfm/")]
			retrieveNfInstanceId = resourceUri[strings.LastIndex(resourceUri, "/")+1:]
			return resouceNrfUri, retrieveNfInstanceId, err
		} else {
			fmt.Println("handler returned wrong status code", status)
			fmt.Println("NRF return wrong status code", status)
		}
	}
}

func SendNFPeriodicHeartbeat(nrfUri, nfInstanceId string, patchItem []models.PatchItem) (string, string, error) {

	// Set client and set url
	configuration := Nnrf_NFManagement.NewConfiguration()
	configuration.SetBasePath(nrfUri)
	client := Nnrf_NFManagement.NewAPIClient(configuration)
	var resouceNrfUri string
	var retrieveNfInstanceId string
	//for {
	_, res, err := client.NFInstanceIDDocumentApi.UpdateNFInstance(context.TODO(), nfInstanceId, patchItem)
	if err != nil || res == nil {
		//TODO : add log
		fmt.Println(fmt.Errorf("NEF UpdateNFInstance to NRF Error[%s]", err.Error()))
		// time.Sleep(2 * time.Second)
		//continue
	} else {
		status := res.StatusCode
		if status == http.StatusOK {
			// NFUpdate
			return resouceNrfUri, retrieveNfInstanceId, err
		} else if status == http.StatusCreated {
			// NFRegister
			resourceUri := res.Header.Get("Location")
			resouceNrfUri = resourceUri[:strings.Index(resourceUri, "/nnrf-nfm/")]
			retrieveNfInstanceId = resourceUri[strings.LastIndex(resourceUri, "/")+1:]
			return resouceNrfUri, retrieveNfInstanceId, err
		} else {
			fmt.Println("handler returned wrong status code", status)
			fmt.Println("NRF return wrong status code", status)
		}
	}
	return resouceNrfUri, retrieveNfInstanceId, err
	//}
}
