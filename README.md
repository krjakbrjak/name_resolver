# Name Resolver

When developing web services, it's often necessary to access endpoints by service name instead of just IP address. While tools like [devdns](https://github.com/ruudud/devdns) exist, they may not work seamlessly across all platforms or can be complex to set up.

**Name Resolver** makes this process simple by running a lightweight DNS server that automatically resolves Docker container names and aliases to their respective IP addresses. If a name doesn't match any running container, it falls back to standard DNS servers (like 1.1.1.1 or 8.8.8.8). This is especially useful for testing containerized applications that require public URLs or service discovery by name.

<p align="center">
  <img src="./flow.svg" alt="Local DNS Container Name Resolution Flow"/>
</p>

## Usage

Build and run locally:
```shell
go build
./name-resolver dns --port 5300
```

Or use Docker:
```shell
docker build . -t name_resolver
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -p 5300:53/udp name_resolver dns
```

> **Note:**
> To use your DNS server, configure your system to point to it. For example, if using `systemd-resolved`:
>
> ```shell
> sudo resolvectl dns <INTERFACE> 127.0.0.1:5300
> ```
>
> This change is temporary and will reset on reboot. To revert manually:
>
> ```shell
> sudo systemctl restart systemd-resolved
> ```

With Name Resolver, you can easily access your containerized services by name, streamlining development and testing.
