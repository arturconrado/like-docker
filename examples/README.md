# Examples

- `rootfs/`: root filesystems de demonstração para `container-linux`.
- `bundles/`: espaço reservado para bundles de runtime/OCI simplificados.

Para preparar um rootfs local:

```bash
make prepare-rootfs
```

Depois configure:

```bash
export MINIDOCK_CONTAINER_ROOTFS=./examples/rootfs/demo
```
