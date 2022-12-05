import os
from ipaddress import IPv4Address, IPv6Address, ip_address
from typing import Dict, List, Optional, Union


class Resolver:
    original: str
    path: str
    data: Dict[str, Union[IPv4Address, IPv6Address]]

    def __init__(self, path: str):
        with open(path, "r") as f:
            self.original = f.read()
            self.path = path
            self.data = {}

    def __setitem__(self, name: str, address: str):
        self.data.update({name: ip_address(address)})
        content: List[str] = []
        with open(self.path, "r") as f:
            content = f.readlines()
        with open(self.path, "w") as f:
            for line in content:
                if name not in line.split():
                    f.write(line)
            f.write(f"{address} {name}\n")

    def __getitem__(self, name: str) -> Optional[Union[IPv4Address, IPv6Address]]:
        return self.data.get(name, None)

    def __len__(self) -> int:
        return len(self.data.items())

    def __contains__(self, item: str):
        return item in self.data

    def __delitem__(self, item: str):
        if item in self:
            content: List[str] = []
            with open(self.path, "r") as f:
                content = f.readlines()
            with open(self.path, "w") as f:
                for line in content:
                    if f"{self[item]} {item}" not in line:
                        f.write(line)
            del self.data[item]

    def __iter__(self):
        return iter(self.data.items())
