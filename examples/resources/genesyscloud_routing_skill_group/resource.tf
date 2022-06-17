resource "genesyscloud_routing_skill_group" "spanish_investment_specialists" {
  name        = "SpanishInvestmentSpecialists"
  description = "Agents with a Series 6 or Series 7 license and are proficient in Spanish (level 3)"
  skill_conditions = jsonencode(
    [
      {
        "routingSkillConditions" : [
          {
            "routingSkill" : "Series 6",
            "comparator" : "EqualTo",
            "proficiency" : 5,
            "childConditions" : {
              "routingSkillConditions" : [
                {
                  "routingSkill" : "Series 7",
                  "comparator" : "EqualTo",
                  "proficiency" : 5,
                  "childConditions" : {
                    "routingSkillConditions" : [],
                    "languageSkillConditions" : [],
                    "operation" : "And"
                  }
                }
              ],
              "languageSkillConditions" : [],
              "operation" : "And"
            }
          }
        ],
        "languageSkillConditions" : [],
        "operation" : "And"
      },
      {
        "routingSkillConditions" : [],
        "languageSkillConditions" : [
          {
            "languageSkill" : "Spanish",
            "comparator" : "GreaterThanOrEqualTo",
            "proficiency" : 3,
            "childConditions" : {
              "routingSkillConditions" : [],
              "languageSkillConditions" : [],
              "operation" : "And"
            }
          }
        ],
        "operation" : "And"
      }
    ]
  )
}