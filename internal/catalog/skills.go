package catalog

import "github.com/Santiago1809/teicos/internal/model"

type Skill struct {
	ID       model.SkillID
	Name     string
	Category string
	Priority string
}

var mvpSkills = []Skill{
	// Foundation skills
	{ID: model.SkillGoTesting, Name: "go-testing", Category: "testing", Priority: "p0"},
	{ID: model.SkillGentleAIBench, Name: "gentle-ai-bench", Category: "testing", Priority: "p0"},
	{ID: model.SkillCreator, Name: "skill-creator", Category: "workflow", Priority: "p0"},
	{ID: model.SkillImprover, Name: "skill-improver", Category: "workflow", Priority: "p0"},
	{ID: model.SkillJudgmentDay, Name: "judgment-day", Category: "workflow", Priority: "p0"},
	{ID: model.SkillBranchPR, Name: "branch-pr", Category: "workflow", Priority: "p0"},
	{ID: model.SkillIssueCreation, Name: "issue-creation", Category: "workflow", Priority: "p0"},
	{ID: model.SkillSkillRegistry, Name: "skill-registry", Category: "workflow", Priority: "p0"},
	// Sustainable review skills
	{ID: model.SkillChainedPR, Name: "chained-pr", Category: "workflow", Priority: "p0"},
	{ID: model.SkillCognitiveDoc, Name: "cognitive-doc-design", Category: "workflow", Priority: "p0"},
	{ID: model.SkillCommentWriter, Name: "comment-writer", Category: "workflow", Priority: "p0"},
	{ID: model.SkillWorkUnitCommits, Name: "work-unit-commits", Category: "workflow", Priority: "p0"},
	{ID: model.SkillSystemicIssueTriage, Name: "systemic-issue-triage", Category: "workflow", Priority: "p0"},
	// Teicos security context layer
	{ID: model.SkillSecurity, Name: "security", Category: "security", Priority: "p0"},
	// Vendored third-party skills
	{ID: model.SkillAPISecurity, Name: "api-security-tester", Category: "security", Priority: "p0"},
	{ID: model.SkillCryptoAudit, Name: "crypto-audit", Category: "security", Priority: "p0"},
	{ID: model.SkillBanditSAST, Name: "bandit-sast", Category: "security", Priority: "p0"},
	{ID: model.SkillSocketSCA, Name: "socket-sca", Category: "security", Priority: "p0"},
	{ID: model.SkillSecurityHeaders, Name: "security-headers-audit", Category: "security", Priority: "p0"},
	{ID: model.SkillSecurityTests, Name: "security-test-generator", Category: "security", Priority: "p0"},
	{ID: model.SkillMobileSecurity, Name: "mobile-security", Category: "security", Priority: "p0"},
	{ID: model.SkillDevSecOps, Name: "devsecops-pipeline", Category: "security", Priority: "p0"},
	{ID: model.SkillDockerScout, Name: "docker-scout-scanner", Category: "security", Priority: "p0"},
	{ID: model.SkillPCIDSS, Name: "pci-dss-audit", Category: "security", Priority: "p0"},
	{ID: model.SkillTDD, Name: "tdd", Category: "testing", Priority: "p0"},
	{ID: model.SkillDiagnosingBugs, Name: "diagnosing-bugs", Category: "testing", Priority: "p0"},
	{ID: model.SkillCodeReview, Name: "code-review", Category: "workflow", Priority: "p0"},
}

func MVPSkills() []Skill {
	skills := make([]Skill, len(mvpSkills))
	copy(skills, mvpSkills)
	return skills
}
