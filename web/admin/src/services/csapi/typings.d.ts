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

  type getAdminToolsetInstancesIdLogParams = {
    /** 实例ID */
    id: string;
    /** 是否获取所有日志，默认只获取最后1MB */
    all?: boolean;
  };

  type getAdminToolsetInstancesIdParams = {
    /** 实例ID */
    id: string;
  };

  type getAdminToolsetInstancesParams = {
    /** 是否获取所有实例，默认只获取运行中的实例 */
    all?: boolean;
    /** 跳过的实例数量，默认0 */
    skip?: number;
    /** 获取的实例数量，默认10 */
    take?: number;
  };

  type getAdminWorkflowsIdLogsNameParams = {
    /** 轮次ID */
    id: number;
    /** 日志名称 */
    name: string;
  };

  type getAdminWorkflowsRoundsIdParams = {
    /** 轮次ID */
    id: number;
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

  type KillToolInstanceReq = {
    signal: number;
  };

  type KillWorkflowJobReq = {
    /** "stop" or "kill" */
    type: string;
  };

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

  type PageDTOModelToolInstanceHistoryDTO = {
    count?: number;
    items?: ToolInstanceHistoryDTO[];
    start?: number;
    total?: number;
  };

  type postAdminToolsetInstancesIdKillParams = {
    /** 实例ID */
    id: string;
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

  type RoundDTO = {
    endTime?: string;
    id?: string;
    startTime?: string;
    tasks?: TaskDTO[];
  };

  type RoundResp = {
    currentRound?: number;
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

  type TaskDTO = {
    args?: string;
    dependencies?: string[];
    description?: string;
    endTime?: string;
    name?: string;
    startTime?: string;
    status?: TaskStatus;
    title?: string;
    type?: string;
  };

  type TaskStatus = 'pending' | 'running' | 'success' | 'failed';

  type ToolArgDTO = {
    default?: any;
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
    allowedSignals?: ToolSignalDTO[];
    args?: ToolArgDTO[];
    description?: string;
    group?: string;
    /** ID is the unique identifier for the toolset. */
    id?: string;
    name?: string;
  };

  type ToolInstanceHistoryDTO = {
    endTime?: string;
    err?: string;
    id?: string;
    isRunning?: boolean;
    launchUserName?: string;
    ret?: number;
    startTime?: string;
    tool?: ToolDTO;
    toolId?: string;
    toolName?: string;
  };

  type ToolSignalDTO = {
    description?: string;
    name?: string;
    value?: number;
  };

  type UpdateWorkflowStatusReq = {
    running: boolean;
  };

  type UserInfoResp = {
    policy?: string[];
    username?: string;
  };
}
