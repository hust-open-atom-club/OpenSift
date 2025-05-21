"use client";
import { ModelResultDto } from "@/service/client";

import { ForkOutlined, GithubOutlined } from "@ant-design/icons";
import { Table, Tag } from "antd";

function DistTypeToText(type: number) {
  const TYPES = [
    "Debian",
    "Arch",
    "Homebrew",
    "Nix",
    "Alpine",
    "Centos",
    "Aur",
    "Deepin",
    "Fedora",
    "Gentoo",
    "Ubuntu",
  ]

  return TYPES[type] || "Unknown";
}

function EcosystemTypeToText(type: number) {
  const TYPES = [
    "Npm",
    "Go",
    "Maven",
    "Pypi",
    "Nuget",
    "Cargo",
    "Others",
  ]

  return TYPES[type] || "Unknown";
}


export function RepoDetail(
  { data }: {
    data: ModelResultDto
  }
) {
  return <div>

    <div className="flex items-center gap-2">
      {data.link?.indexOf("github") !== -1 ? <GithubOutlined /> : <ForkOutlined />}
      <span className="font-bold">{data.link}</span>
    </div>

    <h3 className="my-4 font-bold">Total Score: {data.score}</h3>

    <h3 className="my-4 font-bold">Git Metadata: {data.gitScore}</h3>

    <Table pagination={false} dataSource={data.gitDetail} columns={[
      {
        title: "Commit Frequency",
        dataIndex: "commitFrequency",
      },
      {
        title: "Contributor Count",
        dataIndex: "contributorCount",
      },
      {
        title: "Created Since",
        dataIndex: "createdSince",
        render: (t) => new Date(t).toLocaleString()
      },
      {
        title: "Updated Since",
        dataIndex: "updatedSince",
        render: (t) => new Date(t).toLocaleString()
      },
      {
        title: "Language",
        dataIndex: "language",
        render: (l, r) => r.language?.map(x => <Tag key={x}>{x}</Tag>)
      },
      {
        title: "License",
        dataIndex: "license",
        render: (l, r) => r.license?.map(x => <Tag key={x}>{x}</Tag>)
      },
      {
        title: "Org Count",
        dataIndex: "orgCount",
      },
      {
        title: "Update Time",
        dataIndex: "updateTime",
        render: (t) => new Date(t).toLocaleString()
      }
    ]} />


    <h3 className="my-4 font-bold" >Language Ecosystems: {data.langScore}</h3>
    <Table pagination={false} dataSource={data.langDetail} columns={[
      {
        title: "Type",
        dataIndex: "type",
        render: (t) => EcosystemTypeToText(t)
      },
      {
        title: "Dep Count",
        dataIndex: "depCount",
      },
      {
        title: "Lang Eco Impact",
        dataIndex: "langEcoImpact",
      },
      {
        title: "Update Time",
        dataIndex: "updateTime",
        render: (t) => new Date(t).toLocaleString()
      }
    ]} />

    <h3 className="my-4 font-bold" >Distributions: {data.distroScore}</h3>

    <Table pagination={false} dataSource={data.distDetail} columns={[
      {
        title: "Type",
        dataIndex: "type",
        render: (t) => DistTypeToText(t)
      },
      {
        title: "Count",
        dataIndex: "count",
      },
      {
        title: "Impact",
        dataIndex: "impact",
      },
      {
        title: "Page Rank",
        dataIndex: "pageRank",
      },
      {
        title: "Update Time",
        dataIndex: "updateTime",
        render: (t) => new Date(t).toLocaleString()
      }
    ]} />

  </div>

}