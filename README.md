# vault migrator

[release]: https://github.com/nebtex/vault-migrator/releases

[![GitHub release](http://img.shields.io/github/release/nebtex/vault-migrator.svg?style=flat-square)][release]
[![Go Report Card](https://goreportcard.com/badge/github.com/nebtex/vault-migrator)](https://goreportcard.com/report/github.com/nebtex/vault-migrator)

migrate or backup vault data between two physical backends. in one operation or in a cron job.

tested with: `vault v0.7`, `consul`, `dynamodb`


# Warnings

* Before you run this tool, make sure that you are not running vault in the destination backend

# Usage

Create a `config.json` file with this structure

```json
{
  "to": {
    "name": "[[Backend Name]]",
    "config": "[[Backend Config]]"
  },
    "from": {
        "name": "[[Backend Name]]",
        "config": "{[[Backend Config]]"
    }
}
```

where `from`, is the source backend, and `to`  is the destination

## Examples:

Remember only use strings in the backend config values!!!

from dynamodb to consul

```json
{
  "to": {
    "name": "consul",
      "config": {
        "address": "127.0.0.7:8500",
        "path": "vault",
        "token": "xxxx-xxxx-xxxx-xxxx-xxxxxxxxx"
     }
  },
    "from": {
        "name": "dynamodb",
        "config": {
          "ha_enabled": "true",
          "table": "vault",
          "write_capacity": "1",
          "access_key": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
          "secret_key": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
        }
    },
  "schedule": "@daily"
}
```

this will backup each 24 hours your data in dynamodb to a consul instance. 

full list of storage backends and configuration options: [Vault Storage Backends](https://www.vaultproject.io/docs/configuration/storage/index.html)

`schedule` is optional if is not defined the command will run only once, for more documentation about is format please check [robfig/cron](https://godoc.org/github.com/robfig/cron)


#### Usage

```shell 
vault-migrator --config ${your_config_path}
```

## Building

Use glide and Go 1.10+. Due to Go's package management stupidity, the other step is necessary:

```shell
glide install
rm -rf vendor/github.com/hashicorp/vault/vendor/github.com/hashicorp/go-hclog
```

## Contribution

To contribute to this project, see [CONTRIBUTING](CONTRIBUTING).

## Licensing

*vault-migrator* is licensed under the APACHE License v2. See [LICENSE](LICENSE) for the full license text.

