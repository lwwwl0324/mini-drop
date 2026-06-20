import React from 'react';
import { Layout } from 'tdesign-react';
import 'tdesign-react/es/style/index.css';

const { Header, Content } = Layout;

const AppLayout = ({ children }) => {
  return (
    <Layout style={{ height: '100vh' }}>
      <Header style={{ 
        background: '#1a1a2e', 
        color: 'white', 
        padding: '0 24px',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between'
      }}>
        <div style={{ fontSize: '20px', fontWeight: 'bold' }}>
          🔥 Mini-Drop
        </div>
        <div style={{ fontSize: '14px', color: '#aaa' }}>
          性能分析平台
        </div>
      </Header>
      <Content style={{ padding: '24px', background: '#f5f7fa', overflow: 'auto' }}>
        {children}
      </Content>
    </Layout>
  );
};

export default AppLayout;
