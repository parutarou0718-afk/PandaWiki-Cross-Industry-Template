export const handleThinkingContent = (content: string) => {
  const thinkRegex = /<think>([\s\S]*?)(?:<\/think>|$)/g;
  const thinkMatches = [];
  let match;
  while ((match = thinkRegex.exec(content)) !== null) {
    thinkMatches.push(match[1]);
  }

  let answerContent = content.replace(/<think>[\s\S]*?<\/think>/g, '');
  answerContent = answerContent.replace(/<think>[\s\S]*$/, '');

  return {
    thinkingContent: thinkMatches.join(''),
    answerContent: answerContent,
  };
};
