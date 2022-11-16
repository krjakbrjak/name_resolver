FROM ubuntu:22.04 AS base

RUN apt-get update && apt install -y python3 python3-pip
COPY . .
RUN pip install . -t /tmp/install

FROM ubuntu:22.04

RUN apt-get update && apt install -y python3
COPY --from=base /tmp/install /name_resolver
ENV PYTHONPATH=/name_resolver:$PYTHONPATH
ENV PATH=/name_resolver/bin:$PATH

ENTRYPOINT ["name_resolver"]
