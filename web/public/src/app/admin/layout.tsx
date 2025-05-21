"use client";
import React from 'react';
import { Breadcrumb, ConfigProvider, Layout, Menu, theme } from 'antd';
import { redirect, usePathname } from 'next/navigation';
import zhCN from "antd/locale/zh_CN"
import Link from 'next/link';

const { Header, Content, Footer } = Layout;

const menuItems = [
  { key: "/admin/gitfile", label: "仓库存储" }
]

export default function ({ children }: React.PropsWithChildren<{}>) {
  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken();

  const pathname = usePathname();

  const selectedKeys = menuItems.filter(x => pathname.startsWith(x.key)).map(x => x.key)
  const breadcrumbLabel = menuItems.find(x => pathname.startsWith(x.key))?.label


  return (
    <ConfigProvider theme={{
      components: {
        Layout: {
          headerBg: "white"
        }
      }
    }} locale={zhCN}>
      <Layout>
        <Header className='shadow-md sticky top-0 z-10 w-full flex items-center'>
          <div className='mr-8'>
            <img src='/logo.svg' className='h-8' />
          </div>
          <Menu
            theme="light"
            mode="horizontal"
            items={menuItems.map(x => ({
              ...x,
              label: <Link href={x.key}>{x.label}</Link>
            }))}
            selectedKeys={selectedKeys}
            style={{ flex: 1, minWidth: 0 }}
          />
        </Header>
        <Content style={{ padding: '0 48px' }}>
          <Breadcrumb style={{ margin: '16px 0' }} items={[
            { href: "/admin", title: "管理系统" },
            { href: pathname, title: breadcrumbLabel }
          ]}>
          </Breadcrumb>
          <div
            style={{
              padding: 24,
              minHeight: "calc(100vh - 160px)",
              background: colorBgContainer,
              borderRadius: borderRadiusLG,
            }}
          >
            {children}
          </div>
        </Content>
        <Footer style={{ textAlign: 'center' }}>
          {/* Ant Design ©{new Date().getFullYear()} Created by Ant UED */}
        </Footer>
      </Layout>
    </ConfigProvider>
  );
};
