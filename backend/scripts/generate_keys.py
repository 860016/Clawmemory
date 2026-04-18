from cryptography.hazmat.primitives.asymmetric import rsa
from cryptography.hazmat.primitives import serialization
from pathlib import Path


def generate_rsa_keys(output_dir: str = "keys"):
    Path(output_dir).mkdir(exist_ok=True)
    private_key = rsa.generate_private_key(public_exponent=65537, key_size=2048)
    # Save private key (authorization server only)
    with open(f"{output_dir}/private.pem", "wb") as f:
        f.write(private_key.private_bytes(
            serialization.Encoding.PEM,
            serialization.PrivateFormat.PKCS8,
            serialization.NoEncryption()
        ))
    # Save public key (bundled with OpenClaw)
    public_key = private_key.public_key()
    with open(f"{output_dir}/public.pem", "wb") as f:
        f.write(public_key.public_bytes(
            serialization.Encoding.PEM,
            serialization.PublicFormat.SubjectPKCS1
        ))
    print(f"Keys generated in {output_dir}/")


if __name__ == "__main__":
    generate_rsa_keys()
