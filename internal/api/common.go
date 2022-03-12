package api

const errIDNotValid = "id not valid"
const errInvalidRequest = "invalid request"
const errNotFound = "not found"
const errNotBind = "bind error"
const errUpdate = "update error"
const errCreate = "creation failed"
const errQuery = "query failed"
const errConstraint = "contraint failed"

const r = "repository "
const errRNotFound = r + errNotFound
const errRUpdate = r + errUpdate
const errRCreate = r + errCreate
const errRQuery = r + errQuery
const errRConstraint = r + errConstraint

const bp = "backupPolicy "
const errBPNotFound = bp + errNotFound
const errBPUpdate = bp + errUpdate
const errBPCreate = bp + errCreate
const errBPQuery = bp + errQuery

const bs = "backupSchedule "
const errBSNotFound = bs + errNotFound
const errBSUpdate = bs + errUpdate
const errBSCreate = bs + errCreate
const errBSQuery = bs + errQuery

const ms3p = "backupPolicy "
const errMS3PNotFound = ms3p + errNotFound
const errMS3PUpdate = ms3p + errUpdate
const errMS3PCreate = ms3p + errCreate
const errMS3PQuery = ms3p + errQuery

const b2p = "B2Provider "
const errB2PNotFound = b2p + errNotFound
const errB2PUpdate = b2p + errUpdate
const errB2PCreate = b2p + errCreate
const errB2PQuery = b2p + errQuery

const lp = "localProvider "
const errLPNotFound = lp + errNotFound
const errLPUpdate = lp + errUpdate
const errLPCreate = lp + errCreate
const errLPQuery = lp + errQuery

const awsp = "awsS3Provider "
const errAWSPNotFound = awsp + errNotFound
const errAWSPUpdate = awsp + errUpdate
const errAWSPCreate = awsp + errCreate
const errAWSPQuery = awsp + errQuery

const p = "paths "
const errPNotFound = p + errNotFound
const errPUpdate = p + errUpdate
const errPCreate = p + errCreate
const errPQuery = p + errQuery
