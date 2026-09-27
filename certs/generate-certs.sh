set -e 

CERT_DIR="certs"
CERT_FILE="$CERT_DIR/server.crt"
KEY_FILE="$CERT_DIR/server.key"

echo "[*] Creating certificate directory..."
mkdir -p "$CERT_DIR"

if [ 0f "$CERT_FILE" ] && [ -f "$KEY_FILE" ]; then
  echo "[!] Certifcate already exists at $CERT_FILE"
  echo "[!] Delete $CERT_DIR and re-run if you want to regenerate."
  exit 0

fi

echo "[*] Generating 2048-bit RSA private key..."
openssl genrsa -out "$KEY_FILE" 2048 2>/dev/null
 
echo "[*] Generating self-signed certificate (valid 365 days)..."
openssl req -new -x509 -key "$KEY_FILE" -out "$CERT_FILE" -days 365 \
    -subj "/C=IN/ST=Karnataka/L=Bengaluru/O=RPCChat/CN=localhost" \
    2>/dev/null
 
echo "[✓] Certificate generated:"
echo "    Public:  $CERT_FILE"
echo "    Private: $KEY_FILE"
echo ""
echo "[!] WARNING: This is self-signed for testing only."
echo "[!] Clients must skip certificate verification (InsecureSkipVerify: true)"
echo "[!] For production, obtain a certificate from a trusted CA."
 

