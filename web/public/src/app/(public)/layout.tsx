import clsx from 'clsx';
import styles from './layout.module.css';
import React from 'react';

export default function PublicLayout({ children, modal }: React.PropsWithChildren<{
  modal: React.ReactNode;
}>) {
  return <>
    <div className={clsx(styles['public-container'])}>
      {children}
    </div>
    {modal}
  </>
}