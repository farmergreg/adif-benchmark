FROM golang:1.25.0-trixie

WORKDIR /app
COPY . .
RUN chmod +x benchmark.sh

CMD ["./benchmark.sh"]
