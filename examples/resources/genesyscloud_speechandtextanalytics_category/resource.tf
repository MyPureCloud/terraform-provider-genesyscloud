resource "genesyscloud_speechandtextanalytics_category" "example_category" {
  name             = "Example Category"
  description      = "Example Speech & Text Analytics category"
  interaction_type = "Voice"

  # The top two operand levels (0 and 1) must be "OperandGroup".
  # Leaf operands (level 2+) are "Term" or "Topic".
  # The API requires "inverted" and "occurrence" on every operand.
  criteria = jsonencode({
    type       = "OperandGroup"
    inverted   = false
    occurrence = 1
    operands = [
      {
        type       = "OperandGroup"
        inverted   = false
        occurrence = 1
        operands = [
          {
            type       = "Term"
            inverted   = false
            occurrence = 1
            term = {
              word            = "refund"
              participantType = "External"
            }
          }
        ]
      }
    ]
  })
}
