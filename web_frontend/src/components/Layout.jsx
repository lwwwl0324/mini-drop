import React from 'react';
import { Layout } from 'tdesign-react';
import 'tdesign-react/es/style/index.css';
import { Link, useLocation } from 'react-router-dom';

const { Header, Content } = Layout;

const AppLayout = ({ children }) => {
  const location = useLocation();
  
  const menuItems = [
    { path: '/', label: '📊 任务列表' },
    { path: '/agents', label: '🤖 Agent 列表' },
  ];

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
        <div style={{ display: 'flex', gap: '24px' }}>
          {menuItems.map(item => (
            <Link 
              key={item.path}
              to={item.path}
              style={{ 
                color: location.pathname === item.path ? '#fff' : '#aaa',
                textDecoration: 'none',
                fontSize: '14px',
                padding: '8px 12px',
                borderRadius: '4px',
                background: location.pathname === item.path ? 'rgba(255,255,255,0.1)' : 'transparent'
              }}
            >
              {item.label}
            </Link>
          ))}
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
