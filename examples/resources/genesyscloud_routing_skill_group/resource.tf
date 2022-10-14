resource "genesyscloud_routing_skill_group" "skillgroup" {
  name        = "Series6"
  description = "Agents with exposure to Series 6 license"
  skill_conditions = jsonencode(
    [
      {
        "routingSkillConditions" : [
          {
            "routingSkill" : "Series 6",
            "comparator" : "GreaterThan",
            "proficiency" : 2,
            "childConditions" : [{
              "routingSkillConditions" : [],
              "languageSkillConditions" : [],
              "operation" : "And"
            }]
          }
        ],
        "languageSkillConditions" : [],
        "operation" : "And"
    }]
  )
}