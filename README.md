[![Docker Build Status](https://img.shields.io/docker/build/xperimental/ipromnb.svg?style=flat-square)](https://hub.docker.com/r/xperimental/ipromnb/)

# ipromnb

Kernel for [Jupyter Notebooks](http://jupyter.org/) which can query [Prometheus](https://prometheus.io/) servers.

## Usage

For easy start there is a Docker image: [`xperimental/ipromnb`](https://hub.docker.com/r/xperimental/ipromnb/) which is directly runnable.

There's also a `docker-compose.yml` file in this repo, so manually building and running an image should be as easy as (if you have Docker and docker-compose installed):

```bash
git clone https://github.com/xperimental/ipromnb.git
cd ipromnb
docker-compose up --build
```

This will run a jupyter notebooks instance and map it to port 8888. The directory the repository is in will be mounted as a volume in the "work" directory accessible in the Notebook UI.