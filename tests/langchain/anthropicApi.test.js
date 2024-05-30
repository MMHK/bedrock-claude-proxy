import { ChatAnthropic  } from "@langchain/anthropic";
import { StringOutputParser  } from "@langchain/core/output_parsers";
import { HumanMessage, SystemMessage } from "@langchain/core/messages";

describe('Anthropic API Tests', () => {

    test('normal', async () => {
        const model = new ChatAnthropic({
            temperature: 0.5,
            model: process.env.ANTHROPIC_MODEL,
            topK: 1,
            topP: 1,
        });

        const messages = [
            new SystemMessage("You are a helpful assistant."),
            new HumanMessage("Hello, how are you?"),
        ];

        const response = await model.invoke(messages);

        const parser = new StringOutputParser();

        const result = await parser.invoke(response);

        console.log(result)

        expect(result).toBeDefined();
        expect(result).toContain('Hello');
    });

    test('stream', async () => {
        const model = new ChatAnthropic({
            temperature: 0.5,
            model: process.env.ANTHROPIC_MODEL,
            topK: 1,
            topP: 1,
        });

        const messages = [
            new SystemMessage("You are a helpful assistant."),
            new HumanMessage("Hello, how are you?"),
        ];

        const stream = await model.stream(messages);

        let buffer = '';
        for await (const chunk of stream) {
            console.log(`event: ${chunk._getType()} text: ${chunk.content}`);
            buffer += chunk.content;
        }

        console.log(buffer)

        expect(buffer).toContain('Hello');
    });

    test('image message', async () => {
        const model = new ChatAnthropic({
            temperature: 0.5,
            model: process.env.ANTHROPIC_MODEL,
            topK: 1,
            topP: 1,
        });

        const messages = [
            new SystemMessage("You are a helpful assistant."),
            new HumanMessage({
                content: [
                    {
                        type: 'image_url',
                        image_url: 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAADAAAAAwCAYAAABXAvmHAAAAAXNSR0IArs4c6QAAAgVJREFUaEPtmF1Og0AQgHfAg7Tii2m9gz2J5RaGmLQmhniL1pPIHWzjiygHKawZWmJ/WPZvtpQILyTt7vB9M0MZCqzjB3Scn/UCbVewr0BfAcsM/N8W+rh5usfkXW02GZ5vs9fyrHrgfo8XA1xfgJdhHN0YuFerAusgmnLGHxiDEv7w4AkwePPyImkC2cZgM8ZYCW8SY3+PksDn4HGQ+7CoBz+hyDh44d3XS7L/DWYcePGuViGe+DkPVSoiFdjCez9qF/5bBZzNR9/xM36yvo5mHNhcN4afF0OZRKOAKXwFihIcylaZ6sLv1md+XkyaJBoFVkG0sLi4IfPJtuU4jUNRMKGAbfap6DFOUysJBUz7lhJ8L5awCkKBVRBxRzBGYcdpXMvaGQFRG9UKXFL/V+XSEsBNqyDC3/6ap6VRB1hvMmmhSxLIxmk8rMtC0z1wCc+AillfYHcf4OzSehsBY+EojZdaFbCZYawb/jCAMPu4TDrMtX0zywY6qUCbrcTBmxyP5cfVlQrghjYkVOCVWqgyPqeEKryWwLkqoQOvLeBaQhfeSMCVhAm8sQC1hCm8lQCVhA28tYCthC08iYCpBAU8mYCuBBU8qYCqBCU8uYBMghreiYBIwgW8M4Fqfqr+Qhe9jFC8NyhNoxQXchWjF3CVWdW4fQVUM+VqXV8BV5lVjdv5CvwC+5IGQDHi0L8AAAAASUVORK5CYII=',
                    },
                    {
                        type: 'text',
                        text: 'description of the image' ,
                    },
                ]
            }),
        ];

        const stream = await model.stream(messages);

        let buffer = '';
        for await (const chunk of stream) {
            console.log(`event: ${chunk._getType()} text: ${chunk.content}`);
            buffer += chunk.content;
        }

        console.log(`LLM: \n${buffer}\n`)

        expect(buffer).toContain('heart');
    }, 30000)
});