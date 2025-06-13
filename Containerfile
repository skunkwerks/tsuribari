FROM ghcr.io/freebsd/freebsd-runtime:14.3@sha256:3a5ffe995405b5f6300797b38d87328a267bbeeb550d3707c9c5e0a76827a978

# Install required packages
RUN /bin/pkg install -yr FreeBSD-base FreeBSD-utilities

# Create necessary directories
RUN mkdir -p /usr/local/etc/tsuribari /var/run/tsuribari
RUN chown www:www /usr/local/etc/tsuribari /var/run/tsuribari
RUN chmod 750 /usr/local/etc/tsuribari

# Copy the tsuribari binary
COPY bin/tsuribari /usr/local/bin/tsuribari
RUN chmod +x /usr/local/bin/tsuribari

# Switch to www user
USER www:www

# Expose the default port
EXPOSE 4003

# Set the entrypoint
ENTRYPOINT ["/usr/local/bin/tsuribari"]
