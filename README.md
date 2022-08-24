# postman-ws-proxy v0.1

### postman-ws-proxy saves all recived text messges to file specified
### supports ws and wss and http.Header

# example
```json
{
    // two fields are needed 
    // "ProxyTarget" and "ProxyFileName"
    "ProxyTarget":"wss://example.dev/api/?token=123xyz",
    //or 
    "ProxyTarget":"ws://localhost:58008/api/?token=123xyz",
    // file name to whitch we save recived messages to
    "ProxyFileName":"./logs", // or just 'logs'

    // your original message
    "RequestID": "abc",
    ...
}
```
### log file will be in $HOME/postman-proxy/$ProxyFileName#

# Envs
| Env Name              | default |
| ---------:            |-------|     
| **PP_PORT string**    | 8008 |
| **PP_LOG_LEVEL int8** | 1 |


## log levels same as zerolog
--------------------

DebugLevel defines debug log level.
- DebugLevel 0

InfoLevel defines info log level.
- InfoLevel 1

WarnLevel defines warn log level.
- WarnLevel 2

ErrorLevel defines error log level.
- ErrorLevel 3

FatalLevel defines fatal log level.
- FatalLevel 4

PanicLevel defines panic log level.
- PanicLevel 5

NoLevel defines an absent log level.
- NoLevel 6

Disabled disables the logger.
- Disabled 7