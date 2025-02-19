# Debugging the Genesys Cloud Terraform Provider

This guide outlines the process for setting up and performing live debugging of the Genesys Cloud Terraform Provider source code alongside Terraform Core. More information about debugging plugin providers can be found in the [official docs](https://developer.hashicorp.com/terraform/plugin/debugging).

## Prerequisites

- Go version 1.20.14 (recommended for the Genesys provider)
- Terraform Core
- Debugger: Delve
- Integrated Development Environment (IDE): Visual Studio Code or GoLand
- Sample `main.tf` file with the `terraform` block configured to use the local provider for testing

## Setup

1. Clone the source code:

   ```
   git clone https://github.com/MyPureCloud/terraform-provider-genesyscloud
   ```

2. Install Go 1.20.14 from the [Archived version list](https://go.dev/dl/).

3. Set up your IDE (Visual Studio Code is used in this guide).

## Configuration

1. Create a `launch.json` file in VS Code with the following debug configuration:

   ```json
   {
     "version": "0.2.0",
     "configurations": [
       {
         "name": "Debug Genesys Cloud Terraform Provider",
         "type": "go",
         "request": "launch",
         "mode": "debug",
         "program": "${workspaceFolder}",
         "env": {},
         "args": ["--debug"]
       }
     ]
   }
   ```

2. Build the source code:
   ```
   make build
   make sideload
   ```

## Debugging Process

1. In VS Code, select "Debug Genesys Cloud Terraform Provider" in the Debug view.

2. Start the debugger. This will load the source with the Delve debugger.

3. In the Debug Console, copy the `TF_REATTACH_PROVIDERS` environment variable line.

4. Open a new terminal in VS Code and set the `TF_REATTACH_PROVIDERS` environment variable using the copied line. For example:

   ```
   $ export TF_REATTACH_PROVIDERS='{"genesys.com/mypurecloud/genesyscloud":{"Protocol":"grpc","ProtocolVersion":5,"Pid":94538,"Test":true,"Addr":{"Network":"unix","String":"/var/folders/v2/vr8_wczj64q137jp1rrpvlz40000gp/T/plugin4151936511"}}}'
   ```

5. Set breakpoints in the desired files (e.g., `resource_genesyscloud_routing_wrapupcode.go`).

6. Run `terraform init` followed by `terraform apply` in the terminal.

7. The debugger will pause at the set breakpoints, allowing you to step through the code and inspect variables.

## Notes

- Ensure that the `source` attribute in the `required_providers` block in your `main.tf` file is set to `"genesys.com/mypurecloud/genesyscloud"`
- The `TF_REATTACH_PROVIDERS` environment variable enables debugging with breakpoints.

By following these steps, you can effectively debug the Genesys Cloud Terraform Provider in your development environment.

---

Documentation inspired by [a blog post on the Developer Center](https://developer.genesys.cloud/blog/2024-06-07-debug-cxascode-in-devenv/)
