package gov

import (
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/governmint/types"
	eyesApp "github.com/tendermint/merkleeyes/app"
	eyes "github.com/tendermint/merkleeyes/client"
	"github.com/tendermint/tmsp/server"
	"testing"
)

func makeMerkleEyesServer(addr string) *server.Server {
	app := eyesApp.NewMerkleEyesApp()
	s, err := server.NewServer(addr, app)
	if err != nil {
		panic("starting MerkleEyes listener: " + err.Error())
	}
	return s
}

func makeMerkleEyesClient(addr string) *eyes.Client {
	c, err := eyes.NewClient("unix://test.sock")
	if err != nil {
		panic("creating MerkleEyes client: " + err.Error())
	}
	return c
}

func TestUnit(t *testing.T) {
	s := makeMerkleEyesServer("unix://test.sock")
	defer s.Stop()
	c := makeMerkleEyesClient("unix://test.sock")
	defer c.Stop()
	gov := NewGovernmint(c)

	// Test Entity
	{
		privKey := crypto.GenPrivKeyEd25519()
		pubKey := privKey.PubKey()

		gov.SetEntity(&types.Entity{
			ID:     "my_entity_id",
			PubKey: pubKey,
		})

		entityCopy, ok := gov.GetEntity("my_entity_id")
		if !ok {
			t.Error("Saved(set) entity does not exist")
		}
		if entityCopy.ID != "my_entity_id" {
			t.Error("Got wrong entity id")
		}
		if !pubKey.Equals(entityCopy.PubKey) {
			t.Error("Got wrong entity pubkey")
		}

		entityBad, ok := gov.GetEntity("my_bad_id")
		if ok || entityBad != nil {
			t.Error("Expected nil entity")
		}
	}

	// Test Group
	{
		gov.SetGroup(&types.Group{
			ID:      "my_group_id",
			Version: 1,
			Members: []types.Member{
				types.Member{
					EntityID:    "my_entity_id",
					VotingPower: 1,
				},
			},
		})

		groupCopy, ok := gov.GetGroup("my_group_id")
		if !ok {
			t.Error("Saved(set) group does not exist")
		}
		if groupCopy.ID != "my_group_id" {
			t.Error("Got wrong group id")
		}
		if groupCopy.Version != 1 {
			t.Error("Got wrong group version ")
		}
		if len(groupCopy.Members) != 1 {
			t.Error("Got wrong group members size")
		}
		if groupCopy.Members[0].EntityID != "my_entity_id" {
			t.Error("Group member's entity id is wrong")
		}

		groupBad, ok := gov.GetGroup("my_bad_id")
		if ok || groupBad != nil {
			t.Error("Expected nil group")
		}
	}

	// Test ActiveProposal
	{
		ap := &types.ActiveProposal{
			Proposal: types.Proposal{
				ID: "my_proposal_id",
				Info: &types.GroupUpdateProposalInfo{
					GroupID:     "my_group_id",
					NextVersion: 1,
					ChangedMembers: []types.Member{
						types.Member{
							EntityID:    "entity1",
							VotingPower: 1,
						},
					},
				},
				StartHeight: 99,
				EndHeight:   100,
			},
			SignedVotes: []types.SignedVote{
				types.SignedVote{
					Vote: types.Vote{
						Height:     123,
						EntityID:   "entity1",
						ProposalID: "my_proposal_id",
						Value:      "my_vote",
					},
					Signature: nil, // TODO set a sig
				},
			},
		}
		gov.SetActiveProposal(ap)
		proposalID := ap.Proposal.ID

		apCopy, ok := gov.GetActiveProposal(proposalID)
		if !ok {
			t.Error("Saved(set) ap does not exist")
		}
		if apCopy.Proposal.Info.(*types.GroupUpdateProposalInfo).GroupID != "my_group_id" {
			t.Error("Got wrong ap proposal group id")
		}
		if apCopy.Proposal.Info.(*types.GroupUpdateProposalInfo).NextVersion != 1 {
			t.Error("Got wrong ap proposal group version ")
		}
		if len(apCopy.Proposal.Info.(*types.GroupUpdateProposalInfo).ChangedMembers) != 1 {
			t.Error("Got wrong ap proposal changed members size")
		}
		if len(apCopy.SignedVotes) != 1 {
			t.Error("Got wrong ap proposal votes size")
		}

		apBad, ok := gov.GetActiveProposal("my_bad_id")
		if ok || apBad != nil {
			t.Error("Expected nil ap")
		}
	}
}
