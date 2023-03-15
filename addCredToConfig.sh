if /opt/homebrew/bin/gpg --list-secret-keys | grep 67C54A5E
then
    echo "secret exists"
else
    /opt/homebrew/bin/gpg --no-tty --batch --yes --import secret.asc
    echo "done importing gpg key"
fi

/opt/homebrew/bin/gpg --list-secret-keys
