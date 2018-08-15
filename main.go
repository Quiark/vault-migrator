package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
	"context"

	vaultcommand "github.com/hashicorp/vault/command"
	"github.com/hashicorp/vault/physical"
	log "github.com/hashicorp/go-hclog"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var backendFactories map[string]physical.Factory

func init() {
	// Perform a harmless run to initialize the values.
	vaultcommand.Run([]string{ "version" })

	// fish the backend factories out of the vault CLI, since that is inexplicably where
	// this map is assembled
	vaultCommands := vaultcommand.Commands
	cmd, err := vaultCommands["server"]()
	if err != nil {
		logrus.Fatal("'vault server' init failed", err)
	}
	serverCommand, ok := cmd.(*vaultcommand.ServerCommand)
	if !ok {
		logrus.Fatal("'vault server' did not return a ServerCommand")
	}
	backendFactories = serverCommand.PhysicalBackends
}

func newBackend(kind string, logger log.Logger, conf map[string]string) (physical.Backend, error) {
	if factory := backendFactories[kind]; factory == nil {
		return nil, fmt.Errorf("no Vault backend is named %+q", kind)
	} else {
		return factory(conf, logger)
	}
}

//Backend is a supported storage backend by vault
type Backend struct {
	//Use the same name that is used in the vault config file
	Name string `json:"name"`
	//Put here the configuration of your picked backend
	Config map[string]string `json:"config"`
}

//Config config.json structure
type Config struct {
	//Source
	From *Backend `json:"from"`
	//Destination
	To *Backend `json:"to"`
	//Schedule (optional)
	Schedule *string `json:"schedule"`
}

func moveData(path string, from physical.Backend, to physical.Backend) error {
	ctx := context.Background()
	keys, err := from.List(ctx, path)
	if err != nil {
		return err
	}
	for _, key := range keys {
		logrus.Infoln("moving key: ", path+key)
		if strings.HasSuffix(key, "/") {
			err := moveData(path+key, from, to)
			if err != nil {
				return err
			}
			continue
		}
		entry, err := from.Get(ctx, path + key)
		if err != nil {
			return err
		}
		if entry == nil {
			continue
		}
		err = to.Put(ctx, entry)

		if err != nil {
			return err
		}
	}
	if path == "" {
		logrus.Info("all the keys have been moved ")
	}
	return nil
}

func move(config *Config) error {
	logger := log.New(&log.LoggerOptions{
		Name: "vault-migrator",
	})

	from, err := newBackend(config.From.Name, logger, config.From.Config)
	if err != nil {
		return err
	}
	to, err := newBackend(config.To.Name, logger, config.To.Config)
	if err != nil {
		return err
	}
	return moveData("", from, to)
}

func main() {
	app := cli.NewApp()
	app.Name = "vault-migrator"
	app.Usage = ""
	app.Version = version
	app.Authors = []cli.Author{{"nebtex", "publicdev@nebtex.com"}}
	app.Flags = []cli.Flag{cli.StringFlag{
		Name:   "config, c",
		Value:  "",
		Usage:  "config file",
		EnvVar: "VAULT_MIGRATOR_CONFIG_FILE",
	}}

	app.Action = func(c *cli.Context) error {
		configFile := c.String("config")
		configRaw, err := ioutil.ReadFile(configFile)
		if err != nil {
			return err
		}
		config := &Config{}
		err = json.Unmarshal(configRaw, config)
		if err != nil {
			return err
		}
		if config.From == nil {
			return fmt.Errorf("%v", "Please define a source (key: from)")
		}
		if config.To == nil {
			return fmt.Errorf("%v", "Please define a destination (key: to)")
		}
		if config.Schedule == nil {
			return move(config)
		}
		cr := cron.New()
		err = cr.AddFunc(*config.Schedule, func() {
			defer func() {
				err := recover()
				if err != nil {
					logrus.Errorln(err)
				}
			}()
			err = move(config)
			if err != nil {
				logrus.Errorln(err)
			}
		})
		if err != nil {
			return err
		}
		cr.Start()
		//make initial migration
		err = move(config)
		if err != nil {
			return err
		}
		for {
			time.Sleep(time.Second * 60)
			logrus.Info("Waiting the next schedule")

		}

	}
	err := app.Run(os.Args)
	if err != nil {
		logrus.Fatal(err)
	}
}
