# addons

This repository is the single home for all Harvester addons. Every addon lives
in its own directory under `addons/`:

```
addons/<name>/
├── metadata.yaml        # name, namespace, stage, builtIn, deprecated
├── README.md            # what the addon does and its current status
├── addon-template.yaml  # builtIn addons: fragment of the rancherd bootstrap template
└── addon.yaml           # non-builtIn addons: static manifest, kubectl-apply-able
```

`metadata.yaml` carries two orthogonal fields:

- `stage`: maturity and support commitment — `experimental` (roughly feature
  complete, insufficient test coverage), `preview` (partial detail features and
  partial test coverage) or `ga` (feature complete with sufficient automated
  coverage). The stage determines the `addon.harvesterhci.io/experimental` /
  `addon.harvesterhci.io/preview` / `addon.harvesterhci.io/ga` label on the
  Addon resource.
- `builtIn`: whether the addon is packaged into the Harvester ISO and managed
  by upgrades. This is a usage-driven packaging decision, independent of stage.

`deprecated: true` marks an unmaintained addon (never `builtIn`); see the
per-addon README for its status at capability granularity.

Non-builtIn addons are installed directly from this repository:

```
kubectl apply -f https://raw.githubusercontent.com/harvester/addons/main/addons/<name>/addon.yaml
```

## Generation and validation

The rancherd bootstrap template is assembled from the `builtIn` addon fragments
for harvester-installer packaging as follows:
`go run . -generateTemplates -path $path_to_installer_templates`

For harvester upgrade path, the templates need to be rendered, and easiest way
to do the same is to call

`go run . -generateAddons -path $upgrade_path_manifests`

Both commands first validate the addon layout (metadata invariants, required
files, stage labels); `go run . -validate` runs the checks alone.

The repo also contains a `version_info` file which is sourced by
`harvester-installer` build-bundle script

Please ensure image and chart info update is also reflected in this file.

## run `make` with dapper

All following commands run in similar way like most Harvester repos.

Run `make generate` to generate the addon templates, which is saved under `./bin`.

Run `make test-chart-patch` to test the patches upon `rancher-monitoring` and `rancher-logging` charts, the patched charts are saved under `./bin/patched-charts`.
