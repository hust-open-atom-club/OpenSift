import { getToken } from "@/bearer";
import { useEffect, useRef, useState } from "react";
import { FitAddon } from "@xterm/addon-fit";
import { Terminal } from "@xterm/xterm";
import '@xterm/xterm/css/xterm.css'

type Props = {
  id: string;
  onTruncate?: () => void;
  onRefresh?: () => void;
}

export default function Term({ id, onTruncate, onRefresh }: Props) {

  const ref = useRef<HTMLDivElement>(null);
  const instance = useRef<Terminal>();

  // load xterm
  useEffect(() => {
    if (!ref.current) return;
    const term = new Terminal({
      cursorBlink: true,
      cursorStyle: 'block',
      fontFamily: 'operator mono,SFMono-Regular,Consolas,Liberation Mono,Menlo,monospace',
      fontSize: 14,
      theme: { background: '#101420' },
    });
    instance.current = term;
    const fit = new FitAddon();
    term.loadAddon(fit);
    term.open(ref.current);
    fit.fit();
    const resizeObserver = new ResizeObserver(() => {
      fit.fit();
    });
    resizeObserver.observe(ref.current);

    term.focus();
    return () => {
      term.dispose();
      resizeObserver.disconnect();
    }
  }, [ref]);


  const socket = useRef<WebSocket>();
  const [sizeStr, setSizeStr] = useState("");

  const sendResize = (rows?: number, cols?: number) => {
    if (rows === undefined) rows = instance.current?.rows;
    if (cols === undefined) cols = instance.current?.cols;
    if (rows === undefined || cols === undefined) return;

    setSizeStr(`${cols}x${rows}`);
    // cols, rows (uint16) into big endian
    const d = new Uint8Array(5);
    d[0] = 0x07;
    d[1] = (cols >> 8) & 0xFF;
    d[2] = cols & 0xFF;
    d[3] = (rows >> 8) & 0xFF;
    d[4] = rows & 0xFF;

    console.log("[TERM DEBUG] on resize", cols, rows, d);

    if (socket.current?.readyState === WebSocket.OPEN)
      socket.current?.send(d);

    setTimeout(() => {
      setSizeStr("");
    }, 2000);

  }

  const writeLocalMessage = (data: string) => {
    if (!instance.current) return;
    // write yellow background, black text
    // if current line is not empty, add a new line
    const cursorX = instance.current.buffer.active.cursorX;
    if (cursorX > 0) {
      instance.current.write("\r\n");
    }

    const d = "\x1b[43m\x1b[30m * " + data + "   \x1b[0m\r\n";
    instance.current.write(d);
  }

  useEffect(() => {
    if (!instance.current) return;
    writeLocalMessage(`Connecting to instance ${id} ...`)
    instance.current?.onData((data) => {
      const d = new TextEncoder().encode("\x00" + data);
      // console.log("[TERM DEBUG] on data", d);
      if (socket.current?.readyState === WebSocket.OPEN)
        socket.current?.send(d);
    })
    instance.current?.onBinary((data) => {
      const d = new TextEncoder().encode("\x00" + data);
      // console.log("[TERM DEBUG] on binary", d);
      if (socket.current?.readyState === WebSocket.OPEN)
        socket.current?.send(d);
    })
    instance.current?.onResize(({ cols, rows }) => {
      sendResize(rows, cols);
    });
  }, [instance]);

  useEffect(() => {
    if (!instance) return;
    // websocket
    instance?.current?.reset();
    const ws = new WebSocket(`/api/v1/admin/toolset/instances/${id}/attach?auth_token=${encodeURIComponent(`Bearer ${getToken()}`)}`);
    socket.current = ws;
    socket.current.binaryType = 'arraybuffer';

    const onOpen = () => {
      console.log("[TERM DEBUG] WebSocket connection established");
      sendResize();
    };
    const onError = (e: Event) => {
      console.log("[TERM DEBUG] error", e);
      writeLocalMessage(`Connection error occurred.`);
    };
    const onClose = (e: Event) => {
      console.log("[TERM DEBUG] WebSocket connection closed", e);
      writeLocalMessage(`Connection closed.`);
    }
    const onMessage = (e: MessageEvent) => {
      // if data is Blob
      if (e.data instanceof ArrayBuffer) {
        const d = new Uint8Array(e.data);
        if (d[0] === 0x01 || d[0] === 0x02)
          instance.current?.write(d.slice(1));
        if (d[0] === 0x09) {
          writeLocalMessage(`Instance is not running, dumping logs...`);
          const ww = d.slice(1);
          const truncatedFlag = "truncated.";
          if (ww.slice(0, truncatedFlag.length).toString() === truncatedFlag) {
            writeLocalMessage(`The log is too long, truncated to 1MB.`);
            onTruncate?.();
          }
          instance.current?.write(ww);
        }
        if (d[0] === 0x06) {
          // d[1..5] is return code of big endian
          const code = d[1] << 24 | d[2] << 16 | d[3] << 8 | d[4];
          const errMessage = new TextDecoder().decode(d.slice(5));
          let msg = `Instance exited with code ${code}`
          onRefresh?.();
          if (!!errMessage) {
            msg += `\n${errMessage}`;
          }
          writeLocalMessage(msg);
          ws.close();
          socket.current = undefined;
        }
      } else {
        console.error("[TERM DEBUG] WebSocket message error", e);
      }
    };

    ws.addEventListener('open', onOpen);
    ws.addEventListener('error', onError);
    ws.addEventListener('close', onClose);
    ws.addEventListener('message', onMessage);

    return () => {
      ws.removeEventListener('open', onOpen);
      ws.removeEventListener('error', onError);
      ws.removeEventListener('close', onClose);
      ws.removeEventListener('message', onMessage);
      ws.close();
    }
  }, [id, instance]);

  return <div className="h-full w-full relative">
    {sizeStr &&
      <div className="absolute top-0 right-0 text-gray-500 bg-white text-sm px-4 z-30">
        {sizeStr}
      </div>
    }
    <div ref={ref} className="h-full w-full" />
  </div>
}