/* eslint-disable */
/* tslint:disable */
// @ts-nocheck
/*
 * ---------------------------------------------------------------
 * ## THIS FILE WAS GENERATED VIA SWAGGER-TYPESCRIPT-API        ##
 * ##                                                           ##
 * ## AUTHOR: acacode                                           ##
 * ## SOURCE: https://github.com/acacode/swagger-typescript-api ##
 * ---------------------------------------------------------------
 */

import httpRequest, { ContentType, RequestParams } from "./httpClient";
import {
  DomainPWResponse,
  GetShareV1OpenapiGithubCallbackParams,
  GithubComChaitinPandaWikiApiShareV1GitHubCallbackResp,
  PostShareV1OpenapiLarkBotKbIdParams,
} from "./types";

/**
 * @description GitHub回调
 *
 * @tags ShareOpenapi
 * @name GetShareV1OpenapiGithubCallback
 * @summary GitHub回调
 * @request GET:/share/v1/openapi/github/callback
 * @response `200` `(DomainPWResponse & {
    data?: GithubComChaitinPandaWikiApiShareV1GitHubCallbackResp,

})` OK
 */

export const getShareV1OpenapiGithubCallback = (
  query: GetShareV1OpenapiGithubCallbackParams,
  params: RequestParams = {},
) =>
  httpRequest<
    DomainPWResponse & {
      data?: GithubComChaitinPandaWikiApiShareV1GitHubCallbackResp;
    }
  >({
    path: `/share/v1/openapi/github/callback`,
    method: "GET",
    query: query,
    type: ContentType.Json,
    format: "json",
    ...params,
  });

/**
 * @description Lark机器人请求
 *
 * @tags ShareOpenapi
 * @name PostShareV1OpenapiLarkBotKbId
 * @summary Lark机器人请求
 * @request POST:/share/v1/openapi/lark/bot/{kb_id}
 * @response `200` `DomainPWResponse` OK
 */

export const postShareV1OpenapiLarkBotKbId = (
  { kbId, ...query }: PostShareV1OpenapiLarkBotKbIdParams,
  params: RequestParams = {},
) =>
  httpRequest<DomainPWResponse>({
    path: `/share/v1/openapi/lark/bot/${kbId}`,
    method: "POST",
    type: ContentType.Json,
    format: "json",
    ...params,
  });
