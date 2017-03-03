mirror302
=========

Mirror302 is a web server designed to consume a mirror list and then redirect the user to a valid mirror for their requested resource/path.

The primary use case this is designed to solve is to have a single published URL which fronts a dynamic group of mirrors.  The end users will always point to the single published URL and the request will be handled by one of the established mirrors.

Mirror302 will concurrently check that the requested resource exists with a `HEAD` on each mirror.  The first mirror to respond with a `200` will be chosen to serve the resource and the client will be `302` redirected to the resource on that mirror.


Configuration
-------------

Mirror302 configuration is handled by [Viper](https://github.com/spf13/viper), so the config can be defined in any of JSON, TOML, YAML, HCL, and Java properties config files.

The config file must be named `config.xyz`, where `xyz` is the file extension to match the chosen configuration format from the list above.  The config file must be placed in the same directory as the binary.


**Configuration Options**

- `mirror_list_url` (required) - The complete URL to the mirror list text file.  For example: http://domain.com/path/to/mirrors.txt
- `port` (optional, default: `8080`) - The port the web server runs on.
- `timeout` (optional, default: `10`) - The number of seconds to wait for the HEAD of the requested path before giving up.


**`mirror_list_url`**

This URL references a text file with one mirror URL per line.  The mirrors must include the scheme (HTTP | HTTPS) and can optionally define a path.


Install
-------

Mirror302 is distributed as a single binary (located in `./bin`) which has been cross compiled to be run on all major platforms.  I would recommend running the application using [`supervisord`](http://supervisord.org/).  Here is an example `supervisor.conf` file for your reference.  In this example I am assuming you are running the binary as `your_user` from a `mirror302` directory in your HOME folder.

`supervisor.conf`
```
[program:mirror302]
command=/home/your_user/mirror302/mirror302
autostart=true
autorestart=true
startretries=10
user=your_user
directory=/home/your_user/mirror302/
redirect_stderr=true
stdout_logfile=/var/log/supervisor/mirror302.log
stdout_logfile_maxbytes=50MB
stdout_logfile_backups=10
```


Example
-------

Lets assume the DNS which points at *mirror302* is: `http://download.example.com`

Lets assume the text file referenced by `mirror_list_url` is of the form:
```
http://first.domain.com
http://second.url.com/with/path
http://third.site.com/
```

If an end user attempts to reference a resource with: `http://download.example.com/resource/file.txt`
The request would represent the following resource on each mirror:
```
http://first.domain.com/resource/file.txt
http://second.url.com/with/path/resource/file.txt
http://third.site.com/resource/file.txt
```
