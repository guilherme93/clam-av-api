# Use the official ClamAV image as the base image
FROM clamav/clamav:1.4

# Set environment variables for ClamAV configuration
ENV CLAMD_CONF_MaxFileSize=250M \
    CLAMD_CONF_MaxScanSize=250M

RUN apk add --no-cache ca-certificates

WORKDIR /
COPY file-scan-api /file-scan-api
COPY bootstrap.sh /bootstrap.sh

ENTRYPOINT ["/bin/sh"]
CMD ["bootstrap.sh"]