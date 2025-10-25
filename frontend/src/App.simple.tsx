// import React from 'react'

function App() {
    return (
        <div style={{ padding: '20px', fontFamily: 'Arial, sans-serif' }}>
            <h1 style={{ color: 'blue', fontSize: '24px' }}>CryptoSignal Hunter</h1>
            <p style={{ color: 'green', fontSize: '16px' }}>前端应用正在运行！</p>
            <div style={{
                marginTop: '20px',
                padding: '10px',
                backgroundColor: '#f0f0f0',
                border: '1px solid #ccc',
                borderRadius: '5px'
            }}>
                <p>如果你能看到这个页面，说明React应用已经成功启动。</p>
                <p>当前时间: {new Date().toLocaleString()}</p>
            </div>
        </div>
    )
}

export default App
