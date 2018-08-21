package dockfix

import (
	"io"
	"github.com/sirupsen/logrus"
	"github.com/MeneDev/dockfix/docker_repo"
	"github.com/MeneDev/dockfix/dockfmt"
	"bytes"
	"github.com/MeneDev/dockfix/dockref"
	"github.com/docker/distribution/reference"
	"github.com/pkg/errors"
)

func Find(log logrus.FieldLogger, formatProvider dockfmt.FormatProvider, reader io.Reader, filename string, callback dockfmt.ImageNameProcessor) error {
	log = log.WithFields(logrus.Fields{
		"filename": filename,
	})
	fileFormat, formatError := dockfmt.IdentifyFormat(log, formatProvider, reader, filename)

	if fileFormat == nil {
		log.Info("Unknown Format")
		return formatError
	}

	log.WithFields(logrus.Fields{
		"format": fileFormat,
	}).Debug("Using format")

	var buffer bytes.Buffer
	err := fileFormat.Process(log, reader, &buffer, callback)
	if err != nil {
		return err
	}

	//
	//configFile, err := os.Open(config.Dir() + "/config.json")
	//var objmap map[string]*json.RawMessage
	//configBytes, err := ioutil.ReadAll(configFile)
	//err = json.Unmarshal(configBytes, &objmap)
	//var jsonAuths map[string]map[string]string
	//json.Unmarshal(*objmap["auths"], &jsonAuths)
	//
	//auths := make(map[string]string)
	//for host, auth := range jsonAuths {
	//
	//	bytes, _ := base64.URLEncoding.DecodeString(auth["auth"])
	//	s := string(bytes)
	//	split := strings.Split(s, ":")
	//
	//	authConfig := types.AuthConfig{
	//		Username: split[0],
	//		Password: split[1],
	//	}
	//	encodedJSON, err := json.Marshal(authConfig)
	//	if err != nil {
	//		panic(err)
	//	}
	//	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	//	auths[host] = authStr
	//}
	//
	//
	//println(config.Dir())
	return nil
}

func Pin(log logrus.FieldLogger, formatProvider dockfmt.FormatProvider, repo docker_repo.DockerRepo, reader io.Reader, filename string, w io.Writer) error {
	log = log.WithFields(logrus.Fields{
		"filename": filename,
	})
	fileFormat, formatError := dockfmt.IdentifyFormat(log, formatProvider, reader, filename)

	if fileFormat == nil {
		log.Info("Unknown Format")
		return formatError
	}

	log.WithFields(logrus.Fields{
		"format": fileFormat.Name(),
	}).Debug("Using format")

	var imageNameProcessor dockfmt.ImageNameProcessor = func(r dockref.Reference) (pin string, err error) {
		imageName := r.Name()
		named, err := reference.WithName(imageName)
		named, err = reference.WithTag(named, r.Tag())
		//ref, err := reference.WithDigest(named, r.Digest())

		dig := r.Digest()
		var canonical reference.Reference
		if named != nil && dig != "" {
			canonical, err = reference.WithDigest(named, dig)
		} else {
			if dig != "" {
				canonical, err = repo.FindDigest(dig)
			} else {
				return "", errors.Errorf("No information for " + imageName)
			}
		}


		//print("Verify against registry... ")
		//inspect, err := repo.DistributionInspect(canonical.String())
		//if err != nil {
		//	panic(err)
		//}
		//
		//if inspect.Descriptor.DigestString == dig {
		//	println("OK")
		//} else {
		//	println("FAILED")
		//	panic("Verification against Repository failed")
		//}

		return canonical.String(), nil

		//return "", nil
	}

	err := fileFormat.Process(log, reader, w, imageNameProcessor)
	if err != nil {
		return err
	}

	//
	//configFile, err := os.Open(config.Dir() + "/config.json")
	//var objmap map[string]*json.RawMessage
	//configBytes, err := ioutil.ReadAll(configFile)
	//err = json.Unmarshal(configBytes, &objmap)
	//var jsonAuths map[string]map[string]string
	//json.Unmarshal(*objmap["auths"], &jsonAuths)
	//
	//auths := make(map[string]string)
	//for host, auth := range jsonAuths {
	//
	//	bytes, _ := base64.URLEncoding.DecodeString(auth["auth"])
	//	s := string(bytes)
	//	split := strings.Split(s, ":")
	//
	//	authConfig := types.AuthConfig{
	//		Username: split[0],
	//		Password: split[1],
	//	}
	//	encodedJSON, err := json.Marshal(authConfig)
	//	if err != nil {
	//		panic(err)
	//	}
	//	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	//	auths[host] = authStr
	//}
	//
	//
	//println(config.Dir())
	return nil
}
