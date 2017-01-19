go-smtp-send
============

Simple utility to send an email using an external SMTP server.

Usage
-----

The utility will read the message body (without SMTP headers) from standard input.

A configuration file specifying connection parameters needs to be present in `/etc/go-smtp-send.yaml`, or you must pass the file path using the `-config` flag.


```
$ ./go-smtp-send -h
  -config string
    	config file to use (default "/etc/go-smtp-send.yaml")
  -h	show this help
  -subject string
    	subject of the email
  -to string
    	address to send mail to
```

Example:

```
echo "Hello World" | ./go-smtp-send -to john@example.com -subject "Hello World"
```

Configuration
-------------

An example configuration file is in [go-smtp-send.yaml](go-smtp-send.yaml):

```yaml
server:
  host: smtp.example.com:465
  tls: true
  verify-tls: true
from: noreply@example.com
auth:
  username: noreply@example.com
  password: super-secret-password
```

