import http.server
import socketserver

class CustomHandler(http.server.SimpleHTTPRequestHandler):
    def do_POST(self):
        content_length = int(self.headers['Content-Length'])
        post_data = self.rfile.read(content_length)
        print(post_data.decode() + "\n---\n")
        # Handle the POST data here
        response = b'HTTP/1.1 200 OK\r\n\r\n'
        self.wfile.write(response)

PORT = 8082

with socketserver.TCPServer(("", PORT), CustomHandler) as httpd:
    print("Server running on port", PORT)
    httpd.serve_forever()

