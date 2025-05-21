import { setToken } from '@/bearer';
import { GithubFilled, LogoutOutlined } from '@ant-design/icons';
import { Avatar, Menu, Spin, Space, MenuProps, Dropdown } from 'antd';
import { stringify } from 'querystring';
import type { MenuInfo } from 'rc-menu/lib/interface';
import React, { useCallback } from 'react';
import { history, useModel } from 'umi';

export type GlobalHeaderRightProps = {
  menu?: boolean;
};

/**
 * 退出登录，并且将当前的 url 保存
 */
const logout = async () => {
  setToken('');
  const { search, pathname } = history.location;
  if (window.location.pathname !== '/session') {
    history.replace({
      pathname: '/session',
      search: stringify({
        'ret_uri': pathname + search,
      }),
    });
  }
};

const RightContent: React.FC<GlobalHeaderRightProps> = () => {
  const { initialState, loading } = useModel('@@initialState');


  const loadingEle = (
    <span>
      <Spin
        size="small"
        style={{
          marginLeft: 8,
          marginRight: 8,
        }}
      />
    </span>
  );

  if (!initialState) {
    return loading;
  }

  const { user } = initialState;

  if (loading) {
    return loadingEle;
  }

  const menuItems: MenuProps["items"] = [
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
    },
  ];

  const onMenuClick = (info: MenuInfo) => {
    if (info.key === 'logout') {
      logout();
      return;
    }
  }

  const menuHeaderDropdown = (
    <Menu selectedKeys={[]} onClick={onMenuClick} items={menuItems} />
  );

  return (
    <Dropdown menu={{
      items: menuItems,
      onClick: onMenuClick
    }}>
      <span className='px-4 hover:bg-gray-100 cursor-pointer'>
        <Avatar className='mr-2' size="small" alt="avatar" icon={<GithubFilled />} />
        <Space>
          <span className="anticon">{user?.username}</span>
        </Space>
      </span>
    </Dropdown>
  );
};

export default RightContent;

