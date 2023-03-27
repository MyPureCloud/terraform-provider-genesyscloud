resource "genesyscloud_knowledge_document_v1" "example_document" {
  knowledge_base_id = genesyscloud_knowledge.example_knowledgebase.id
  language_code     = "en-US"
  knowledge_document {
    type         = "Faq"
    external_url = "https://www.example.com/"
    faq {
      question = "What are the 4 pillars of OOP?"
      alternatives = [
        "What are the pillars of OOP",
        "What are the four pillars of object oriented programming?",
        "What are the principles of object oriented programming?",
      ]
      answer = "Abstraction, Encapsulation, Inheritance, Polymorphism"
    }
  }
}