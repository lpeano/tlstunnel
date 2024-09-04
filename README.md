# tlstunnel
Make tnls tunnel autenticated from client Certificate for application not able to do it

## Configuration
Configuration must be placed in same folder of executable with name config.yaml

It contains following values in yaml format

      
      loglevel: ERROR
      # Do not change value of poolsize!
      # Not yet supported
      poolsize: 1
      tlscfg:
        certfile: C:\path\to\public.crt
        keyfile: C:\path\to\privat.key
        proxyhost: host:port
      netcfg:
        keepalive: 2s
        deadline: 30s

You require the client certificate to connect to gateway.

After this you can copy your executable and config in same folder and run them.

## Kubernete client

For kubernete client you have to define the path 127.0.0.1:8080 as proxy for HTTPS and HTTP like follow example in .kube\config:
