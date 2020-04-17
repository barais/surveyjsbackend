var exam = {
    "title" : "Examen SIR",
    "questions": [
      {
        "name": "photo",
        "type": "photocapture",
        "title": "Prenez une jolie photo que l' on sache que c'est vous",
        "isRequired:": true
    },
   {
            "name": "birthdate",
            "type": "text",
            "inputType": "date",
            "title": "Quel est la date du jour:",
            "isRequired": true
        },         {
          "type": "file",
          "title": "Importer un fichier doc contenant votre rapport",
          "name": "rapport",
          "storeDataAsText": true,
          "showPreview": false,
          "maxSize": 102400
      }, {
            "name": "email",
            "type": "text",
            "inputType": "email",
            "title": "Votre e-mail:",
            "placeHolder": "jon.snow@nightwatch.org",
            "isRequired": true,
            "validators": [
                {
                    "type": "email"
                }
            ]
        },
        {
          "type": "radiogroup",
          "hasOther": false,
          "isRequired": true,
          "name": "favoritePet",
          "title": "What is your favorite pet ![A parrot](https://surveyjs.io/Content/Images/examples/markdown/image_16x16.svg =16x16) ?",
          "choices": [
              {
                  "value": "dog",
                  "text": "Dog: ![A dog](https://surveyjs.io/Content/Images/examples/markdown/dog.svg =14x14)"
              }, {
                  "value": "cat",
                  "text": "Cat: ![A cat](https://surveyjs.io/Content/Images/examples/markdown/cat.svg =14x14)"
              }, {
                  "value": "parrot",
                  "text": "Parrot ![A parrot](https://surveyjs.io/Content/Images/examples/markdown/parrot.svg =14x14)"
              }
          ]
      },
      {
        "type": "radiogroup",
        "hasOther": false,
        "isRequired": true,
        "name": "estcedujava",
        "title": "Lequel est du Java valide:",
        "choices": [
       {
        "value": "js",
        "text": `\`\`\`js
function sayHello (msg, who) {
    return \`\${who} says: msg\`;
}
sayHello("Hello World", "Johnny");
\`\`\``
      },
      {
        "value": "java",
        "text": `\`\`\`java
class Foo{
  String a;
  public void bar(){
    System.out.println(a);
  }
}
\`\`\``
      },
      {
        "value": "go",
        "text": `\`\`\`go
package main
import "fmt"
    func main() {
        fmt.Println("hello world")
    }
\`\`\``
      }
    ]
    },


        {
            "type": "editor",
            "name": "feedbackue",
            "title": "En quoi cet UE était interessante"

        },
        {
          "name": 'uml',
          "type": 'uml',
          "title": 'Faites un joli diagram uml',
          "isRequired": true
      },

    ],
    completedHtml: "<p><h4>Merci pour avoir compléter ce projet</h4></p>"

}
console.log(JSON.stringify(exam));