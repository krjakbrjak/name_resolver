# Name resolver

When developing the web services it is important to be able to reach the published endpoints by the service name not just IP address. There are some exisiting projects/tools that help one do that. For example, [devdns](https://github.com/ruudud/devdns). But it does not work on all platforms, or it takes a lot of time and effort to make it work. The easier approach is just to edit `/etc/hosts` file (`C:\Windows\System32\drivers\etc\hosts` for Windows). This project takes [devdns](https://github.com/ruudud/devdns) as a base and edit the `hosts` file dynamically when a new docker container starts/stops.

## Usage

```shell
python3 -m pip install .
name_resolver
```

or start docker container:
```shell
docker build . -t name_resolver
docker run -v /etc/hosts:/etc/hosts -v /var/run/docker.sock:/var/run/docker.sock --rm name_resolver
```
