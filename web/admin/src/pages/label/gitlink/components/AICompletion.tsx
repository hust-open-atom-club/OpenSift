import { Button } from "antd";
import { getToken } from "@/bearer";
import React, { useEffect } from "react";
import Markdown from "react-markdown";
import styles from "./AICompletion.module.css"
import { GoogleOutlined } from "@ant-design/icons";

type Props = {
  data?: API.DistributionPackageDTO;
  distribution?: string;
}

export default function AICompletion({ data, distribution }: Props) {

  const [msg, setMsg] = React.useState<any[]>([]);
  const [loading, setLoading] = React.useState<boolean>(false);

  useEffect(() => {
    setMsg([]);
  }, [data, distribution]);

  const addLine = (text: string) => {
    if (!text) return;
    if (!text.startsWith("data: ")) return;
    const content = text.substring(6).trim();
    const obj = JSON.parse(content);

    setMsg((prev) => {
      return [...prev, obj];
    });
  }

  const startAskAI = (async () => {
    if (!distribution || !data?.package) return;
    setLoading(true);
    try {
      const d = await fetch('/api/v1/admin/label/distributions/ai-completion', {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${getToken()}`,
        },
        method: 'POST',
        body: JSON.stringify({
          distribution: distribution,
          packageName: data.package,
          description: data?.description,
          homePage: data?.homePage,
        }),
      })

      const reader = d.body?.getReader();
      if (!reader) {
        console.error("Failed to get reader from response body");
        return;
      }
      const decoder = new TextDecoder();
      let buf = "";
      while (true) {
        const { value, done } = await reader.read()
        if (done) {
          if (buf) {
            addLine(buf);
            buf = "";
          }
          console.log("Stream finished");
          break;
        }

        buf += decoder.decode(value, { stream: true });
        const lines = buf.split('\n\n');
        buf = lines.pop() || ""; // Keep the last incomplete line in buf
        for (const line of lines) {
          if (line.trim() === "") continue; // Skip empty lines
          addLine(line);
        }
      }
    } finally {
      setLoading(false);
    }
  });

  const text = msg.flatMap((m, i) => {
    return m?.candidates?.flatMap((c: any) => {
      return c?.content?.parts?.flatMap((p: any) => {
        return p?.text || "";
      })
    })
  }).join("") || "";

  console.log(msg)

  const querys: string[] = msg.flatMap((m) => {
    return m?.candidates?.flatMap((c: any) => {
      return c?.groundingMetadata?.webSearchQueries;
    })
  }).filter(x => !!x) || [];

  const groundLinks: {
    title: string;
    uri: string;
  }[] = msg.flatMap((m) => {
    return m?.candidates?.flatMap((c: any) => {
      return c?.groundingMetadata?.groundingChunks?.flatMap((g: any) => {
        return {
          title: g?.web?.title || "",
          uri: g?.web?.uri || ""
        }
      });
    });
  }).filter(g => !!g);

  const tokenUsed = msg?.[msg.length - 1]?.usageMetadata?.totalTokenCount;

  return (
    <div className="my-2">
      <Button loading={loading} type="primary" onClick={startAskAI}>
        {
          msg.length > 0 ? "重新生成" : "询问 AI"
        }
      </Button>
      {msg && msg.length > 0 && <div className={styles['completion']}>
        {text && <div className={styles['markdown-body']}>
          <div className="text-xs text-gray-500 mb-2">
            以下内容由 Gemini AI 生成，可能包含错误或不准确的信息，请谨慎使用
          </div>
          <Markdown >
            {text}
          </Markdown>
        </div>}
        {groundLinks.length > 0 && <div className={styles['gound-links']}>
          <h3>相关链接</h3>
          <ol>
            {groundLinks.map((link, index) => (
              <li key={index}>
                <a href={link.uri} target="_blank" rel="noopener noreferrer">
                  {link.title}
                </a>
              </li>
            ))}
          </ol>
        </div>}
        {querys.length > 0 && <div className={styles['search-queries']}>
          <h3>搜索查询</h3>
          <div>
            {querys.map((query, index) => (
              <div key={index} >
                <GoogleOutlined className="mr-1" />
                <a href={`https://www.google.com/search?q=${encodeURIComponent(query)}`} target="_blank" rel="noopener noreferrer">
                  {query}
                </a>
              </div>
            ))}
          </div>
        </div>}

        {tokenUsed && <div className="text-right text-xs text-gray-500 mt-2">
          当前共消耗了 {tokenUsed} 个 token
        </div>}
      </div>}
    </div>
  )

}