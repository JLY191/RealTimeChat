import React, { useEffect, useState } from 'react';

const Chat = () => {
    const [messages, setMessages] = useState([]);
    const [inputMessage, setInputMessage] = useState('');
    const [socket, setSocket] = useState(null);

    useEffect(() => {
        // 确保只在组件加载时创建一次 WebSocket 对象
        const newSocket = new WebSocket('ws://localhost:12345/ws');

        // 监听连接打开事件
        newSocket.onopen = () => {
            console.log('WebSocket连接已打开');
            setSocket(newSocket); // 更新 state 中的 socket 引用
        };

        // 监听消息接收事件
        newSocket.onmessage = (event) => {
            const message = JSON.parse(event.data);
            setMessages((prevMessages) => [...prevMessages, message]);
        };

        // 监听连接关闭事件
        newSocket.onclose = () => {
            console.log('WebSocket连接已关闭');
            setSocket(null); // 清除之前的 socket 引用
        };

        // 在组件卸载时关闭 WebSocket 连接
        return () => {
            newSocket.close();
        };
    }, []); // 空数组表示只在组件加载时执行

    // 发送消息的函数
    const sendMessage = () => {
        if (socket && inputMessage) {
            socket.send(inputMessage);
            setInputMessage('');
        }
    };

    return (
        <div>
            <ul>
                {messages.map((message, index) => (
                    <li key={index}>{message.sender + ": "}{message.content}</li>
                ))}
            </ul>
            <input
                type="text"
                value={inputMessage}
                onChange={(e) => setInputMessage(e.target.value)}
            />
            <button onClick={sendMessage}>发送消息</button>
        </div>
    );

};

export default Chat;
