variable "run_id" {
  description = "Unique suffix to avoid resource name collisions between test runs."
  type        = string
  default     = "001"
}

# Row count and column mix are fixed at generate time (default 8000 rows, 30 columns).
# To change them, edit generate.py and run: python3 generate.py
