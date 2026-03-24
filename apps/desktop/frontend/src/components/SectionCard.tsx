import { PropsWithChildren } from 'react';

export function SectionCard({ title, subtitle, className, children }: PropsWithChildren<{ title: string; subtitle?: string; className?: string }>) {
  return (
    <section className={className ? `section-card ${className}` : 'section-card'}>
      <div className="section-header">
        <div>
          <h3>{title}</h3>
          {subtitle ? <p>{subtitle}</p> : null}
        </div>
      </div>
      {children}
    </section>
  );
}
