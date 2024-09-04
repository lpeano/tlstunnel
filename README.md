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
        certfile: C:\Users\lpeano\projects\Bozze\tlstunnel\peano.crt
        keyfile: C:\Users\lpeano\projects\Bozze\tlstunnel\peano.key
        proxyhost: host:port
      netcfg:
        keepalive: 2s
        deadline: 30s
