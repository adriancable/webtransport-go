# webtransport-go

This package provides a lightweight WebTransport-over-HTTP/3 server implementation in Go.

## What is WebTransport?

WebTransport (https://www.w3.org/TR/webtransport/) is a 21st century replacement for WebSockets. It's
currently supported by Chrome, with support in other browsers coming shortly. It has several benefits
over WebSockets:

- It's built on top of HTTP/3 which uses QUIC (not TCP) as the transport. QUIC provides significant performance advantages over TCP, especially on congested networks.

- Unlike WebSockets where you get just a single send/receive stream, WebTransport provides for multiple server-initiated or client-initiated unidirectional or bidirectional
(reliable) streams, plus (unreliable) datagrams.

- Unlike XHR/fetch calls, WebTransport provides a means (serverCertificateHashes) for a browser to communicate securely with endpoints where issuance of
Web PKI certificates is not possible, for example, a device over the LAN which is only accessible via a private IP address and hence a self-signed certificate. This is potentially big for IoT.

Neither WebTransport nor HTTP/3 are standardized yet. We adhere to: [draft-ietf-quic-http-34](https://datatracker.ietf.org/doc/html/draft-ietf-quic-http-34) and [draft-ietf-webtrans-http3-02](https://datatracker.ietf.org/doc/html/draft-ietf-webtrans-http3-02).

## Complete Go server and browser-based client example

Take a look at the [webtransport-go-example](https://github.com/adriancable/webtransport-go-example) repo for a complete server (and browser client) example.

## Minimal "getting started" example

You'll need to get a certificate for your server. Please read the comments at the top of [Google's WebTransport server example in Python](https://github.com/GoogleChrome/samples/blob/gh-pages/webtransport/webtransport_server.py) for detailed instructions.

First, set up WebTransport server parameters:

```
server := &webtransport.Server{
	ListenAddr:     ":4433",
	TLSCert:        webtransport.CertFile{Path: "cert.pem"},
	TLSKey:         webtransport.CertFile{Path: "cert.key"},
	AllowedOrigins: []string{"googlechrome.github.io", "127.0.0.1:8000", "localhost:8000"},
}
```

Then, set up an `http.Handler` to accept a session, wait for an incoming bidirectional stream from the client, then (in this example) receive data and echo it back:

```
http.HandleFunc("/counter", func(rw http.ResponseWriter, r *http.Request) {
	session := r.Body.(*webtransport.Session)
	session.AcceptSession()
	defer session.CloseSession()

	// Wait for incoming bidi stream
	s, err := session.AcceptStream()
	if err != nil {
		return
	}

	for {
		buf := make([]byte, 1024)
		n, err := s.Read(buf)
		if err != nil {
			break
		}
		fmt.Printf("Received from bidi stream %v: %s\n", s.StreamID(), buf[:n])
		sendMsg := bytes.ToUpper(buf[:n])
		fmt.Printf("Sending to bidi stream %v: %s\n", s.StreamID(), sendMsg)
		s.Write(sendMsg)
	}
}
```

Finally, start the server:

```
ctx, cancel := context.WithCancel(context.Background())
server.Run(ctx)
```

Here is a simple Chrome browser client to talk to this server. You can open a new browser tab and paste it into the Chrome DevTools console:

```
let transport = new WebTransport("https://localhost:4433/counter");
await transport.ready;
let stream = await transport.createBidirectionalStream();
let encoder = new TextEncoder();
let decoder = new TextDecoder();
let writer = stream.writable.getWriter();
let reader = stream.readable.getReader();
await writer.write(encoder.encode("Hello, world!"))
console.log(decoder.decode((await reader.read()).value));
transport.close();
```

## Authors and acknowledgment
This package was written by me, drawing on several sources of inspiration:

- [Experimenting with QUIC and WebTransport in Go](https://centrifugal.github.io/centrifugo/blog/quic_web_transport/) which describes the implementation in Go of a simple server for WebTransport over QUIC, which is now defunct but formed the basis for the WebTransport over HTTP/3 draft standard.
- [Pull request #3256](https://github.com/lucas-clemente/quic-go/pull/3256) for the `quic-go` package, which implements (an earlier draft spec for) WebTransport but is a heavyweight structural change for the package for which there doesn't appear any appetite from the package maintainer.
- HTTP/3 server built into [`quic-go`](https://github.com/lucas-clemente/quic-go). This package doesn't use `quic-go`'s HTTP/3 server (only the underlying QUIC transport implementation), but `webtransport-go` implements its own minimal HTTP/3 server as needed to support WebTransport.

Instead, `webtransport-go` implements the latest draft standards for all the supported protocols, is very lightweight, and doesn't depend on any non-standard Go package forks or WIP modules.

## License
Provided under the MIT license.
