if gpg --list-secret-keys | grep 25D753B7C560659B057B714C970A8360B4BF5075
then
    echo "secret exists"
else
    gpg --no-tty --batch --yes --import /tmp/terraform_gpg_secret.asc
    echo "done importing gpg key"
fi

gpg --list-secret-keys
