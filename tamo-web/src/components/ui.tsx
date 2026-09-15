'use client';

import { ButtonHTMLAttributes, InputHTMLAttributes, ReactNode, SelectHTMLAttributes, TextareaHTMLAttributes } from 'react';

export function Button({ variant = 'primary', busy = false, children, disabled, ...props }: ButtonHTMLAttributes<HTMLButtonElement> & { variant?: 'primary' | 'secondary' | 'ghost' | 'danger'; busy?: boolean }) {
  return <button className={`ui-button ui-button--${variant}`} disabled={disabled || busy} aria-busy={busy || undefined} {...props}>{busy ? 'Please wait…' : children}</button>;
}

export function TextField({ label, error, ...props }: InputHTMLAttributes<HTMLInputElement> & { label: string; error?: string }) {
  const id = props.id ?? `field-${label.toLowerCase().replace(/[^a-z0-9]+/g, '-')}`;
  return <label className="ui-field" htmlFor={id}><span>{label}</span><input id={id} aria-invalid={Boolean(error)} aria-describedby={error ? `${id}-error` : undefined} {...props}/>{error && <small id={`${id}-error`} role="alert">{error}</small>}</label>;
}

export function SelectField({ label, children, ...props }: SelectHTMLAttributes<HTMLSelectElement> & { label: string; children: ReactNode }) {
  const id = props.id ?? `select-${label.toLowerCase().replace(/[^a-z0-9]+/g, '-')}`;
  return <label className="ui-field" htmlFor={id}><span>{label}</span><select id={id} {...props}>{children}</select></label>;
}

export function TextareaField({ label, error, ...props }: TextareaHTMLAttributes<HTMLTextAreaElement> & { label: string; error?: string }) {
  const id = props.id ?? `textarea-${label.toLowerCase().replace(/[^a-z0-9]+/g, '-')}`;
  return <label className="ui-field" htmlFor={id}><span>{label}</span><textarea id={id} aria-invalid={Boolean(error)} aria-describedby={error ? `${id}-error` : undefined} {...props}/>{error && <small id={`${id}-error`} role="alert">{error}</small>}</label>;
}

export function Choice({ type, label, ...props }: InputHTMLAttributes<HTMLInputElement> & { type: 'checkbox' | 'radio'; label: string }) {
  return <label className="ui-choice"><input type={type} {...props}/><span>{label}</span></label>;
}

export function Toggle({ label, ...props }: InputHTMLAttributes<HTMLInputElement> & { label: string }) {
  return <label className="ui-toggle"><input type="checkbox" role="switch" {...props}/><span aria-hidden="true"/><b>{label}</b></label>;
}

export function Card({ children, className = '' }: { children: ReactNode; className?: string }) { return <section className={`ui-card ${className}`.trim()}>{children}</section>; }
export function EmptyState({ title, message, action }: { title: string; message: string; action?: ReactNode }) { return <div className="empty-state" role="status"><h2>{title}</h2><p>{message}</p>{action}</div>; }
export function Toast({ kind = 'success', children }: { kind?: 'success' | 'error' | 'info'; children: ReactNode }) { return <div className={`ui-toast ui-toast--${kind}`} role={kind === 'error' ? 'alert' : 'status'}>{children}</div>; }

export function Modal({ open, title, children, onClose }: { open: boolean; title: string; children: ReactNode; onClose: () => void }) {
  if (!open) return null;
  return <div className="ui-modal-backdrop" role="presentation" onMouseDown={onClose}><section className="ui-modal" role="dialog" aria-modal="true" aria-labelledby="modal-title" onMouseDown={(event)=>event.stopPropagation()}><header><h2 id="modal-title">{title}</h2><Button variant="ghost" aria-label="Close dialog" onClick={onClose}>×</Button></header>{children}</section></div>;
}

export function Dropdown({ label, children }: { label: string; children: ReactNode }) { return <details className="ui-dropdown"><summary>{label}</summary><div role="menu">{children}</div></details>; }

export function Pagination({ page, pages, onPage }: { page: number; pages: number; onPage: (page: number) => void }) {
  return <nav className="pagination" aria-label="Pagination"><Button variant="ghost" disabled={page <= 1} onClick={()=>onPage(page-1)}>Previous</Button>{Array.from({length:pages},(_,index)=>index+1).map(value=><Button variant="ghost" aria-current={value===page?'page':undefined} key={value} onClick={()=>onPage(value)}>{value}</Button>)}<Button variant="ghost" disabled={page >= pages} onClick={()=>onPage(page+1)}>Next</Button></nav>;
}
