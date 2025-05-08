resource "genesyscloud_knowledge_document_variation" "example_document_variation" {
  knowledge_base_id     = genesyscloud_knowledge_knowledgebase.example_knowledgebase.id
  knowledge_document_id = genesyscloud_knowledge_document.example_unpublished_document.id
  published             = true
  knowledge_document_variation {
    body {
      blocks {
        type = "Paragraph"
        paragraph {
          blocks {
            type = "Text"
            text {
              text      = "Paragraph text"
              marks     = ["Bold", "Italic", "Underline"]
              hyperlink = "https://example.com/hyperlink"
            }
          }
          blocks {
            type = "Image"
            image {
              url       = "https://example.com/image"
              hyperlink = "https://example.com/hyperlink"
            }
          }
        }
      }
      blocks {
        type = "Video"
        video {
          url = "https://example.com/video"
        }
      }
      blocks {
        type = "UnorderedList"
        list {
          blocks {
            type = "ListItem"
            blocks {
              type = "Text"
              text {
                text = "List item"
              }
            }
          }
        }
      }
      blocks {
        type = "Image"
        image {
          url       = "https://example.com/image"
          hyperlink = "https://example.com/hyperlink"
        }
      }
    }
  }
}
