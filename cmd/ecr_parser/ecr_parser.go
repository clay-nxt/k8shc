
package ecr_parser

import "strings"

func ParseImage(image string) (string, string, string) {
    tag := "latest"
    repo := ""
    name := ""

    slashParts := strings.Split(image, "/")
    if len(slashParts) >= 2 {
        repo = slashParts[0]
        imagePart := strings.Join(slashParts[1:], "/")

        colonParts := strings.Split(imagePart, ":")
        if len(colonParts) == 2 {
            name = colonParts[0]
            tag = colonParts[1]
        } else {
            name = imagePart
        }
    } else {
        colonParts := strings.Split(image, ":")
        if len(colonParts) == 2 {
            name = colonParts[0]
            tag = colonParts[1]
        } else {
            name = image
        }
    }

    return repo, name, tag
}
