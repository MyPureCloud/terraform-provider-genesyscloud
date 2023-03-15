if gpg --list-secret-keys | grep CD8C59D
then
    echo "secret exists"
else
    gpg --no-tty --batch --yes --import /tmp/terraform_gpg_secret.asc
    echo "done importing gpg key"
fi

gpg --list-secret-keys
