declare namespace API {
  type getAdminGitfilesParams = {
    /** Git link */
    link?: string;
    /** Filter, 0: no filter, 1: success, 2: fail, 3: never success */
    filter?: number;
    /** Skip count */
    skip?: number;
    /** Take count */
    take?: number;
  };

  type getAdminSessionGithubCallbackParams = {
    /** GitHub OAuth Code */
    code: string;
  };

  type getAdminToolsetInstancesIdAttachParams = {
    /** 实例ID */
    id: string;
  };

  type getAdminToolsetInstancesIdLogParams = {
    /** 实例ID */
    id: string;
  };

  type getHistoriesParams = {
    /** Git link */
    link: string;
    /** Skip count */
    start?: number;
    /** Take count */
    take?: number;
  };

  type getRankingsParams = {
    /** Skip count */
    start?: number;
    /** Take count */
    take?: number;
    /** Include details */
    detail?: boolean;
  };

  type getResultsParams = {
    /** Search query */
    q: string;
    /** Skip count */
    start?: number;
    /** Take count */
    take?: number;
  };

  type getResultsScoreidParams = {
    /** Score ID */
    scoreid: number;
  };

  type GitFileAppendManualReq = {
    gitLink?: string;
  };

  type GitFileDTO = {
    failedTimes?: number;
    filePath?: string;
    gitLink?: string;
    lastSuccess?: string;
    message?: string;
    success?: boolean;
    takeStorage?: number;
    takeTimeMs?: number;
    updateTime?: string;
  };

  type GitFileStatisticsResultDTO = {
    fail?: number;
    neverSuccess?: number;
    success?: number;
    total?: number;
  };

  type GitFileStatusResp = {
    collector?: StatusResp;
    gitFile?: GitFileStatisticsResultDTO;
  };

  type GitHubCallbackResp = {
    token?: string;
  };

  type GitHubClientIDResp = {
    clientId?: string;
  };

  type H = true;

  type PageDTOModelGitFileDTO = {
    count?: number;
    items?: GitFileDTO[];
    start?: number;
    total?: number;
  };

  type PageDTOModelRankingResultDTO = {
    count?: number;
    items?: RankingResultDTO[];
    start?: number;
    total?: number;
  };

  type PageDTOModelResultDTO = {
    count?: number;
    items?: ResultDTO[];
    start?: number;
    total?: number;
  };

  type RankingResultDTO = {
    distDetail?: ResultDistDetailDTO[];
    distroScore?: number;
    gitDetail?: ResultGitMetadataDTO[];
    gitScore?: number;
    langDetail?: ResultLangDetailDTO[];
    langScore?: number;
    link?: string;
    ranking?: number;
    score?: number;
    scoreID?: number;
    updateTime?: string;
  };

  type ResultDistDetailDTO = {
    count?: number;
    impact?: number;
    pageRank?: number;
    type?: number;
    updateTime?: string;
  };

  type ResultDTO = {
    distDetail?: ResultDistDetailDTO[];
    distroScore?: number;
    gitDetail?: ResultGitMetadataDTO[];
    gitScore?: number;
    langDetail?: ResultLangDetailDTO[];
    langScore?: number;
    link?: string;
    score?: number;
    scoreID?: number;
    updateTime?: string;
  };

  type ResultGitMetadataDTO = {
    commitFrequency?: number;
    contributorCount?: number;
    createdSince?: string;
    language?: string[];
    license?: string[];
    orgCount?: number;
    updateTime?: string;
    updatedSince?: string;
  };

  type ResultLangDetailDTO = {
    depCount?: number;
    langEcoImpact?: number;
    type?: number;
    updateTime?: string;
  };

  type RunningTaskDTO = {
    link?: string;
    progress?: string;
    start?: string;
  };

  type StatusResp = {
    currentTasks?: RunningTaskDTO[];
    isRunning?: boolean;
    pendingTasks?: string[];
  };

  type ToolArgDTO = {
    default?: string;
    description?: string;
    /** Name is the name of the argument. */
    name?: string;
    type?: string;
  };

  type ToolCreateInstanceReq = {
    args?: Record<string, any>;
    toolId?: string;
  };

  type ToolDTO = {
    args?: ToolArgDTO[];
    description?: string;
    /** ID is the unique identifier for the toolset. */
    id?: string;
    name?: string;
  };

  type ToolInstanceDTO = {
    id?: string;
    startTime?: string;
    tool?: ToolDTO;
  };

  type UserInfoResp = {
    policy?: string[];
    username?: string;
  };
}
