## Cub

A small tool to restore/backup Consul KV data to a local file as json format

## Build and Install

```bash
$ go get github.com/jsvisa/cub
$ cd $GOPATH/src/github.com/jsvisa/cub
$ go install
```

## Backup from Consul

```bash
$ $GOPATH/bin/cub -addr http://127.0.0.1:8500 -backup -path foo
```

## Restore to Consul

```bash
$ $GOPATH/bin/cub -addr http://127.0.0.1:8500 -restore -path foo
```
