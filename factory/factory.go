/*
 * nef Configuration Factory
 */

package factory

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"

	"github.com/geekaamit/NEF-service/logger"
)

var NefConfig Config

func checkErr(err error) {
	if err != nil {
		err = fmt.Errorf("[Configuration] %s", err.Error())
		logger.AppLog.Fatal(err)
	}
}

// TODO: Support configuration update from REST api
func InitConfigFactory(f string) error {
	if content, err := ioutil.ReadFile(f); err != nil {
		return err
	} else {
		NefConfig = Config{}

		if yamlErr := yaml.Unmarshal(content, &NefConfig); yamlErr != nil {
			return yamlErr
		}
	}

	return nil
}

/*
func InitConfigFactory(f string) {
	content, err := ioutil.ReadFile(f)
	checkErr(err)

	NefConfig = Config{}

	err = yaml.Unmarshal([]byte(content), &NefConfig)
	checkErr(err)

	logger.InitLog.Infof("Successfully initialize configuration %s", f)
}
*/

func CheckConfigVersion() error {
	currentVersion := NefConfig.GetVersion()

	if currentVersion != NEF_EXPECTED_CONFIG_VERSION {
		return fmt.Errorf("config version is [%s], but expected is [%s].",
			currentVersion, NEF_EXPECTED_CONFIG_VERSION)
	}

	logger.CfgLog.Infof("config version [%s]", currentVersion)

	return nil
}
