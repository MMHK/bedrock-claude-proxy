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
});