import * as API from "api/api"
import { createMachine, assign } from "xstate"
import * as TypesGen from "api/typesGenerated"

export const workspaceAgentLogsMachine = createMachine(
  {
    predictableActionArguments: true,
    id: "workspaceAgentLogsMachine",
    schema: {
      events: {} as {
        type: "ADD_STARTUP_LOG"
        log: TypesGen.WorkspaceAgentStartupLog
      },
      context: {} as {
        agentID: string
        startupLogs?: TypesGen.WorkspaceAgentStartupLog[]
      },
      services: {} as {
        getStartupLogs: {
          data: TypesGen.WorkspaceAgentStartupLog[]
        }
      },
    },
    tsTypes: {} as import("./workspaceAgentLogsXService.typegen").Typegen0,
    initial: "loading",
    states: {
      loading: {
        invoke: {
          src: "getStartupLogs",
          onDone: {
            target: "watchStartupLogs",
            actions: ["assignStartupLogs"],
          },
        },
      },
      watchStartupLogs: {
        id: "watchingStartupLogs",
        invoke: {
          id: "streamStartupLogs",
          src: "streamStartupLogs",
        },
      },
      loaded: {
        type: "final",
      },
    },
    on: {
      ADD_STARTUP_LOG: {
        actions: "addStartupLog",
      },
    },
  },
  {
    services: {
      getStartupLogs: (ctx) => API.getWorkspaceAgentStartupLogs(ctx.agentID),
      streamStartupLogs: (ctx) => async (callback) => {
        return new Promise<void>((resolve, reject) => {
          const proto = location.protocol === "https:" ? "wss:" : "ws:"
          let after = 0
          if (ctx.startupLogs && ctx.startupLogs.length > 0) {
            after = ctx.startupLogs[ctx.startupLogs.length - 1].id
          }
          const socket = new WebSocket(
            `${proto}//${location.host}/api/v2/workspaceagents/${ctx.agentID}/startup-logs?follow&after=${after}`,
          )
          socket.binaryType = "blob"
          socket.addEventListener("message", (event) => {
            callback({ type: "ADD_STARTUP_LOG", log: JSON.parse(event.data) })
          })
          socket.addEventListener("error", () => {
            reject(new Error("socket errored"))
          })
          socket.addEventListener("open", () => {
            resolve()
          })
        })
      },
    },
    actions: {
      assignStartupLogs: assign({
        startupLogs: (_, { data }) => data,
      }),
      addStartupLog: assign({
        startupLogs: (context, event) => {
          const previousLogs = context.startupLogs ?? []
          return [...previousLogs, event.log]
        },
      }),
    },
  },
)
