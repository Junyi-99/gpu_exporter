# gpu_exporter
A fixed version of gpu_exporter that fetching nvidia-smi.

## Usage
```bash
go build gpu_exporter.go
```

the compiled file will be your current working directory.

## Notes:

The default listening port is `9101`

The metrics url is `/metrics`
