# Migration note

> [!IMPORTANT]
> `kubesort` has been migrated to [codeberg.org/xrstf/kubesort](https://codeberg.org/xrstf/kubesort).

---

## kubesort – Sort Kubernetes YAML files

kubesort is a command-line application that can sort a list of Kubernetes manifests (expressed as
YAML files) in a consistent way. The main usecase is for diffing the output of for example rendered
Helm charts to easier see actual differences between versions.

kubesort will sort manifests by GVK, namespace and name, plus has a number of rules to sort fields
inside of manifests (for example, the environment variables in a PodSpec are sorted by name, so
are containers and volumes).

### Installation

Either [download the latest release](https://github.com/xrstf/kubesort/releases) or build for
yourself using Go 1.22+:

```bash
go install go.xrstf.de/kubesort
```

### Usage

Couldn't really be any simpler:

```bash
Usage of kubesort:
  -c, --config string   Load configuration from this file
  -f, --flatten         Unwrap List kinds into standalone objects
  -V, --version         Show version info and exit immediately
```

Either run kubesort by giving any number of files as arguments:

```bash
$ kubesort deployments.yaml rbac.yaml policies.yaml config.yaml
```

This will combine all the manifests into one, then sort it and return the combined output.

Alternatively, pipe YAML into kubesort on stdin.

### License

MIT
