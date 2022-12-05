import signal
import sys
from enum import Enum
from logging import Logger

import docker
from pydantic import BaseModel, Field

from name_resolver.logger import create_logger
from name_resolver.resolver import Resolver


class Action(str, Enum):
    attach = "attach"
    commit = "commit"
    copy = "copy"
    create = "create"
    destroy = "destroy"
    detach = "detach"
    die = "die"
    exec_create = "exec_create"
    exec_detach = "exec_detach"
    exec_die = "exec_die"
    exec_start = "exec_start"
    export = "export"
    health_status = "health_status"
    kill = "kill"
    oom = "oom"
    pause = "pause"
    rename = "rename"
    resize = "resize"
    restart = "restart"
    start = "start"
    stop = "stop"
    top = "top"
    unpause = "unpause"
    update = "update"


class Attributes(BaseModel):
    name: str


class Actor(BaseModel):
    ID: str
    attributes: Attributes = Field(None, alias="Attributes")


class ContainerEvent(BaseModel):
    action: Action = Field(None, alias="Action")
    actor: Actor = Field(None, alias="Actor")

    class Config:
        use_enum_values = True


def main(resolver: Resolver, logger: Logger):
    client = docker.from_env()

    for container in client.containers.list():
        for _, network in container.attrs["NetworkSettings"]["Networks"].items():
            resolver[container.name] = network["IPAddress"]
            logger.info(f"Added entry: {container.name} -> {resolver[container.name]}")

    for event in client.events(
        decode=True,
        filters={"type": "container", "event": [Action.start, Action.die]},
    ):
        container_event = ContainerEvent(**event)
        if container_event.action == Action.die:
            del resolver[container_event.actor.attributes.name]
            logger.info(f"Removed entry: {container_event.actor.attributes.name}")
        elif container_event.action == Action.start:
            container = client.containers.get(container_event.actor.ID)
            for _, network in container.attrs["NetworkSettings"]["Networks"].items():
                try:
                    resolver[container.name] = network["IPAddress"]
                    logger.info(
                        f"Added entry: {container.name} -> {resolver[network['IPAddress']]}"
                    )
                    logger.info(
                        f"Added entry: {container.name} -> {resolver[container.name]}"
                    )
                except ValueError:
                    logger.error(
                        f"Invalid entry: {container.name} -> {resolver[container.name]}"
                        + " (missing IP-address)"
                    )


def entry():
    resolver = Resolver("/etc/hosts")
    logger = create_logger("Name resolver")

    def handler(signum, frame):
        signame = signal.Signals(signum).name
        delete = [key for key, _ in resolver]
        for i in delete:
            del resolver[i]
            logger.error(f"Removed entry: {i}")
        sys.exit(0)

    signal.signal(signal.SIGINT, handler)
    signal.signal(signal.SIGTERM, handler)

    main(resolver, logger)


if __name__ == "__main__":
    entry()
