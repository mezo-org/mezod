# Upstream Pyth v10.5.0 bundles unpatched Node 24.12.0, and no Pyth release
# up to v13.0.0 has a patched one. Swap in a patched node binary from the
# same major line so the prebuilt native modules stay ABI-compatible.
# Both images are Alpine/musl. Used by pyth-pusher-{mezo,ethereum} only.
#
#   make build-docker-pyth-price-pusher
#   docker push mezo/pyth-price-pusher:v10.5.0-node24.22.1
ARG PYTH_TAG
ARG NODE_TAG
FROM node:${NODE_TAG} AS node

FROM public.ecr.aws/pyth-network/xc-price-pusher:${PYTH_TAG}
USER root
COPY --from=node /usr/local/bin/node /usr/local/bin/node
RUN apk --no-cache upgrade
USER pyth
RUN node --version
