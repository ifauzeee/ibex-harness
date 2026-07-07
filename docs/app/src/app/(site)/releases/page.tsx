import type { Metadata } from "next";

import { readChangelogReleases } from "@/lib/changelog";

const versionBadge = {
  major:
    "rounded-full bg-text-primary px-2.5 py-0.5 font-mono text-xs font-bold text-canvas",
  minor:
    "rounded-full border border-border bg-panel px-2.5 py-0.5 font-mono text-xs font-bold text-text-primary",
  patch:
    "rounded-full border border-border bg-panel-raised px-2.5 py-0.5 font-mono text-xs font-bold text-text-secondary",
} as const;

export const metadata: Metadata = {
  title: "Changelog",
  description: "Version history and release notes generated from docs/CHANGELOG.md.",
};

export default function ReleasesPage() {
  const allReleases = readChangelogReleases();

  return (
    <div className="mx-auto max-w-2xl px-4 py-12 md:px-6 md:py-16 lg:px-8">
      <header className="mb-10 border-b border-border pb-8">
        <p className="mb-3 text-xs font-semibold uppercase tracking-widest text-text-tertiary">
          Releases
        </p>
        <h1 className="mb-3 text-4xl font-bold tracking-tight text-text-primary">
          Changelog
        </h1>
        <p className="text-lg text-text-secondary">
          Full version history and release notes for IBEX Harness.
        </p>
      </header>

      <div className="relative">
        {allReleases.length === 0 ? (
          <div className="rounded-lg border border-border bg-panel p-6 text-sm text-text-secondary">
            No tagged releases yet. Release entries appear here automatically
            when Release Please updates `docs/CHANGELOG.md` for the first
            published version.
          </div>
        ) : null}

        {allReleases.length > 1 ? (
          <div className="absolute bottom-0 left-[11px] top-2 w-px bg-border" />
        ) : null}

        <div className="space-y-0">
          {allReleases.map((release) => {
            const releaseType = release.type;
            const badgeClass = versionBadge[releaseType] ?? versionBadge.minor;

            return (
              <div
                key={release.version}
                className="relative flex gap-6 pb-10 last:pb-0"
              >
                <div className="relative z-10 mt-1 shrink-0">
                  <div className="flex size-6 items-center justify-center rounded-full border-2 border-border bg-canvas">
                    <div className="size-2 rounded-full bg-text-primary" />
                  </div>
                </div>

                <div className="min-w-0 flex-1">
                  <div className="mb-2 flex flex-wrap items-center gap-3">
                    <span className="font-mono text-xl font-bold tracking-tight text-text-primary">
                      v{release.version}
                    </span>
                    <span className={badgeClass}>{releaseType}</span>
                    {release.date ? (
                      <time className="text-sm text-text-secondary">
                        {new Date(release.date).toLocaleDateString("en-US", {
                          year: "numeric",
                          month: "long",
                          day: "numeric",
                        })}
                      </time>
                    ) : null}
                  </div>

                  {release.summary ? (
                    <p className="mb-3 text-base text-text-secondary">
                      {release.summary}
                    </p>
                  ) : null}

                  <div className="space-y-4 text-sm leading-relaxed text-text-secondary">
                    {release.sections
                      .filter((section) => section.items.length > 0)
                      .map((section) => (
                      <section key={`${release.version}-${section.title}`}>
                        <h2 className="mb-2 text-base font-semibold text-text-primary">
                          {section.title}
                        </h2>
                        <ul className="list-disc space-y-1 pl-5">
                          {section.items.map((item) => (
                            <li key={item}>{item}</li>
                          ))}
                        </ul>
                      </section>
                    ))}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
