# Z-Edge Mainflux

Modified instance of a [mainflux/mainflux](https://github.com/mainflux/mainflux) instance.

# Modifications

* Uses [EMQ X Broker](https://github.com/emqx/emqx) instead of [Eclipse Mosquitto](https://github.com/eclipse/mosquitto)

# Setup

```bash
git clone https://github.com/michaelkargl/mainflux.git -b z-edge mainflux_z-edge
cd mainflux_michaelkargl

# This is a fork of a fork => setup both upstream references to update properly
git remote add upstream https://github.com/mteodor/mainflux.git
git remote add upstream_root https://github.com/mainflux/mainflux
```

# Update

```bash
# Check if upstream has changes committed to its raspberry branch
git pull --rebase upstream raspberry
# Update software with root repositories' master feature
git pull --rebase upstream_root master
```
