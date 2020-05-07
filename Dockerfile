FROM scratch
COPY dist/hail-hydra /hail-hydra

ENTRYPOINT ["/hail-hydra"]
