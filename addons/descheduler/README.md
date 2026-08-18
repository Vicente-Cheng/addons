# descheduler

Automatically rebalances virtual machine workloads across nodes (displayed as
"virtual-machine-auto-balance" in the UI), based on the Kubernetes descheduler
with a Harvester-specific eviction policy targeting virt-launcher pods.

- Stage: **experimental** (insufficient automated test coverage yet)
- Built-in: **yes** (shipped in the ISO, disabled by default)

Documentation: https://docs.harvesterhci.io/latest/advanced/addons/virtual-machine-auto-balance
