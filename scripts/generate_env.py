## This script is included in our .devcontainers setup.  We basically use this to generate our an environment variables file that can be used to set the Genesys Cloud environment variables.

import os

def get_user_input(prompt):
    return input(prompt).strip()

def generate_shell_script(client_id, client_secret, region):
    script_content = f"""#!/bin/bash

export GENESYSCLOUD_OAUTHCLIENT_ID="{client_id}"
export GENESYSCLOUD_OAUTHCLIENT_SECRET="{client_secret}"
export GENESYSCLOUD_REGION="{region}"

echo "Environment variables set successfully!"
"""
    return script_content

def main():
    print("Please enter the following values to generate the environment variables needed to run terraform against your target environment: ")
    
    client_id = get_user_input("GENESYSCLOUD_OAUTHCLIENT_ID: ")
    client_secret = get_user_input("GENESYSCLOUD_OAUTHCLIENT_SECRET: ")
    region = get_user_input("GENESYSCLOUD_REGION: ")

    script_content = generate_shell_script(client_id, client_secret, region)

    script_filename = "/tmp/set_genesys_env.sh"
    with open(script_filename, "w") as f:
        f.write(script_content)

    print(f"\nShell script '{script_filename}' has been generated.")
    print(f"To use it, run: source {script_filename}")

    #Make the script executable
    os.chmod(script_filename, 0o755)

if __name__ == "__main__":
    main()