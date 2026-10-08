---
id: ADR-0036
title: 管理画面承認ビューと Bootstrap マスタ拡張の採用
status: Accepted
scope: System
primary_category: APP
categories: [APP, DOM, DEV]
tags: [local-first, single-source-of-truth, yagni]
deciders: [user, ai]
date: 2026-10-08
---

# 0036. 管理画面承認ビューと Bootstrap マスタ拡張の採用 (0036-admin-gallery-and-bootstrap-master-extension.md)

- **ステータス**: 承認 (Accepted)
- **日付**: 2026-10-08
- **関連 ADR**:
  - [ADR-0004: Go 構造体を唯一の Single Source of Truth とする型定義一元管理](./0004-go-schema-as-single-source-of-truth.md)
  - [ADR-0006: プレフィックス付きサロゲートキーと動的ドメインスキーマ構成](./0006-domain-schema-and-typeid-structure.md)
  - [ADR-0007: Local-First / オフライン PWA](./0007-local-first-offline-pwa-architecture.md)
  - [ADR-0017: Repository インターフェース集約と Pure Go SQLite](./0017-repository-interface-and-pure-go-sqlite-architecture.md)
  - [ADR-0021: GitOps マスタデータ同期](./0021-gitops-master-data-synchronization-and-versioning-architecture.md)
  - [ADR-0035: メタデータ確認・編集 API と GitOps 逆同期](./0035-metadata-verification-and-editing-architecture.md)

---

## 1. 背景と解決すべき課題 (Context & Problem)

1. **管理画面での卒業生表示と既存フィルターの不整合**:
   - note14 Session 2 は、グループ・期生・ステータスによるメンバー絞り込みを要求している。
   - フロントエンドには `includeGraduated` フィルターが存在するが、Bootstrap の取得元である `ListMembers` が `status != 'graduated'` を固定しているため、卒業生がクライアントへ届かず、フィルターが機能しない。
2. **PhotoType マスタの取得不足**:
   - Bootstrap は各メンバーの登録画像と紐づく `PhotoType` を返すが、画像に未使用のPhotoTypeを含むマスタ一覧は返していない。
   - 管理画面で未登録の衣装種別へタグを付け替えるには、動的に管理されるPhotoType一覧が必要である ([ADR-0006](./0006-domain-schema-and-typeid-structure.md), [ADR-0021](./0021-gitops-master-data-synchronization-and-versioning-architecture.md))。
3. **管理画面専用参照APIの重複リスク**:
   - 既存BootstrapはLocal-Firstのマスタ同期経路であり、管理画面の一覧表示に必要なメンバー・画像情報をすでに提供している ([ADR-0007](./0007-local-first-offline-pwa-architecture.md))。
   - 同じ情報のために管理画面専用の一覧APIを新設すると、取得経路、キャッシュ、型、エラー処理の二重管理が発生する。

## 2. 決定事項と具現化仕様 (Decision & Specification)

1. **[Frontend] ユーザー向け編集提案画面**:
   - ペンライト色・左右順序・期生・ステータス・代表写真・衣装タグの変更は、公開マスタを直接更新せず、ADR-0037の編集提案として送信する。
2. **[Frontend] `/admin` 承認ビュー**:
   - `/admin` は編集案の変更前後を確認し、承認または却下するための専用ビューとする。メタデータ編集操作はユーザー向け画面に配置する。
3. **[Backend] 承認済みマスタへの反映**:
   - 承認時に編集案の状態変更と公開マスタへの反映を同一トランザクションで行う。提案の保存・状態遷移・競合検出は ADR-0037 に委譲する。
4. **[Contract] Bootstrap の既存経路拡張**:
   - 既存 `GET /api/v1/sync/bootstrap` に `include_graduated` オプションを追加し、`BootstrapResponse` に全 `photo_types` を追加する。管理画面専用の参照APIは新設しない。
5. **[Tool] 手動スナップショットからの逆同期**:
   - 初期は整合性のあるSQLiteスナップショットを手動取得し、承認済みマスタをシードへ決定的に出力する。取得処理と変換処理を分離し、Kubernetes Jobによる自動化は別ADRで扱う。
6. **[Backend] 公開マスタの同期保証**:
   - 承認による公開マスタ更新をBootstrapのETagに反映し、クライアントが最新の承認済みデータを取得できるようにする。取得条件の違いもETag計算に含める。
7. **[Security] 初期運用の保護範囲**:
   - 初期は `/admin` と管理APIの認証・認可保護を行わず、承認機能を先に運用する。Cloudflare Access等による保護は別ADRで扱う。

## 3. 採否の根拠とトレードオフ (Rationale & Trade-offs)

| 選択肢 | 評価 |
|---|---|
| ADR-0035の直接更新方式 | 即時反映は速いが、ユーザー編集を管理者が確認できず、承認履歴も残せない |
| 提案テーブル＋同一SQLiteで承認反映 | 採用。追加DBなしで変更前後・状態・競合・冪等性を管理できる |
| 管理用の別DB・承認キュー | 分離は強いが、DB間同期と運用負荷が増える |

ユーザー編集を提案として保存し、承認済みの内容だけを公開マスタへ反映する。ADR-0035の直接更新方式は、ADR-0036/0037の採択後に置き換える。既存のGitOps逆同期は、承認済みマスタを対象とする部分を継承する。

初期は`/admin`と管理APIを保護しないため、承認者の本人性を保証できない。この運用上の制約は明記し、認証・認可の導入は別ADRで扱う。

## 4. 得られる効果と運用規約 (Consequences & Enforcement)

1. **[Backend]** 未承認・却下の提案をBootstrapとシード逆同期へ混入させない。
2. **[Backend]** 承認反映は編集案の状態更新・マスタ更新・リビジョン更新を同一トランザクションで行う。
3. **[Tool]** SQLiteスナップショットからのシード出力は決定的にし、同じ入力の再実行で不要な差分を生成しない。初期のスナップショット取得は手動とし、Kubernetes Jobは別ADRで扱う。
4. **[Contract]** BootstrapのETagは承認済みマスタの更新と取得条件を反映し、クライアントが古い公開データを継続利用しないようにする。
5. **[Security]** `/admin`保護前は、URLを知る利用者が承認操作できる前提を運用規約に記録する。Cloudflare Access等の保護導入を別ADRで決定する。
6. **[Verification]** 提案の再送・再承認・同時承認・競合・途中失敗をテストし、二重反映と部分更新がないことを確認する。
