package actions_test

import (
	"errors"
	"time"

	dockerContainer "github.com/docker/docker/api/types/container"
	dockerImage "github.com/docker/docker/api/types/image"
	"github.com/docker/go-connections/nat"
	"github.com/fugginold/dockwatch/internal/actions"
	"github.com/fugginold/dockwatch/pkg/container"
	"github.com/fugginold/dockwatch/pkg/types"

	. "github.com/fugginold/dockwatch/internal/actions/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func getCommonTestData(keepContainer string) *TestData {
	return &TestData{
		NameOfContainerToKeep: keepContainer,
		Containers: []types.Container{
			CreateMockContainer(
				"test-container-01",
				"test-container-01",
				"fake-image:latest",
				time.Now().AddDate(0, 0, -1)),
			CreateMockContainer(
				"test-container-02",
				"test-container-02",
				"fake-image:latest",
				time.Now()),
			CreateMockContainer(
				"test-container-02",
				"test-container-02",
				"fake-image:latest",
				time.Now()),
		},
	}
}

func getLinkedTestData(withImageInfo bool) *TestData {
	staleContainer := CreateMockContainer(
		"test-container-01",
		"/test-container-01",
		"fake-image1:latest",
		time.Now().AddDate(0, 0, -1))

	var imageInfo *dockerImage.InspectResponse
	if withImageInfo {
		imageInfo = CreateMockImageInfo("test-container-02")
	}
	linkingContainer := CreateMockContainerWithLinks(
		"test-container-02",
		"/test-container-02",
		"fake-image2:latest",
		time.Now(),
		[]string{staleContainer.Name()},
		imageInfo)

	return &TestData{
		Staleness: map[string]bool{linkingContainer.Name(): false},
		Containers: []types.Container{
			staleContainer,
			linkingContainer,
		},
	}
}

var _ = Describe("the update action", func() {
	// Whether a container may be started again is a fact about that container, not
	// about its image. Tracking it per image means one container sharing an image
	// with another decides the outcome for both -- and the container that failed to
	// stop is the one dockwatch must leave alone, since it is still running the
	// config it was refused permission to replace.
	When("one of two containers sharing an image cannot be stopped", func() {
		It("should not start the container it failed to stop", func() {
			testData := getCommonTestData("test-container-01")
			client := CreateMockClient(testData, false, false)

			_, err := actions.Update(client, types.UpdateParams{})
			Expect(err).NotTo(HaveOccurred())

			Expect(testData.StartedContainers).NotTo(ContainElement("test-container-01"),
				"a container that could not be stopped must not be started")
		})
	})

	When("dockwatch has been instructed to clean up", func() {
		When("there are multiple containers using the same image", func() {
			It("should only try to remove the image once", func() {
				client := CreateMockClient(getCommonTestData(""), false, false)
				_, err := actions.Update(client, types.UpdateParams{Cleanup: true})
				Expect(err).NotTo(HaveOccurred())
				Expect(client.TestData.TriedToRemoveImageCount).To(Equal(1))
			})
		})
		When("there are multiple containers using different images", func() {
			It("should try to remove each of them", func() {
				testData := getCommonTestData("")
				testData.Containers = append(
					testData.Containers,
					CreateMockContainer(
						"unique-test-container",
						"unique-test-container",
						"unique-fake-image:latest",
						time.Now(),
					),
				)
				client := CreateMockClient(testData, false, false)
				_, err := actions.Update(client, types.UpdateParams{Cleanup: true})
				Expect(err).NotTo(HaveOccurred())
				Expect(client.TestData.TriedToRemoveImageCount).To(Equal(2))
			})
		})
		When("there are linked containers being updated", func() {
			It("should not try to remove their images", func() {
				client := CreateMockClient(getLinkedTestData(true), false, false)
				_, err := actions.Update(client, types.UpdateParams{Cleanup: true})
				Expect(err).NotTo(HaveOccurred())
				Expect(client.TestData.TriedToRemoveImageCount).To(Equal(1))
			})
		})
		When("performing a rolling restart update", func() {
			It("should try to remove the image once", func() {
				client := CreateMockClient(getCommonTestData(""), false, false)
				_, err := actions.Update(client, types.UpdateParams{Cleanup: true, RollingRestart: true})
				Expect(err).NotTo(HaveOccurred())
				Expect(client.TestData.TriedToRemoveImageCount).To(Equal(1))
			})
		})
		// Update marks every container that is not monitor-only, including ones whose
		// staleness check already failed. Losing the skip there is how an unreachable
		// registry came to be reported as "everything is up to date".
		When("a container's staleness check fails", func() {
			It("should report it as skipped rather than fresh", func() {
				data := getCommonTestData("")
				data.StalenessError = map[string]error{
					"test-container-01": errors.New("pull failed: unauthorized"),
				}
				client := CreateMockClient(data, false, false)

				report, err := actions.Update(client, types.UpdateParams{})
				Expect(err).NotTo(HaveOccurred())

				Expect(report.Skipped()).To(HaveLen(1))
				Expect(report.Skipped()[0].Name()).To(Equal("test-container-01"))
				Expect(report.Skipped()[0].Error()).To(ContainSubstring("unauthorized"))
				for _, c := range report.Fresh() {
					Expect(c.Name()).NotTo(Equal("test-container-01"))
				}
			})
		})
		When("updating a linked container with missing image info", func() {
			It("should gracefully fail", func() {
				client := CreateMockClient(getLinkedTestData(false), false, false)

				report, err := actions.Update(client, types.UpdateParams{})
				Expect(err).NotTo(HaveOccurred())
				// VerifyConfiguration fails for the linked container, so stopStaleContainer
				// returns before StopContainer (update.go:155-160): it is deliberately left
				// running its original image rather than being stopped with no way back.
				//
				// It still belongs in Failed, not Fresh. The update it was queued for did
				// not happen, so calling it fresh asserts it is up to date when it is
				// stale and currently un-updatable -- and notifications read these buckets.
				// This reverses an earlier deliberate choice to leave it out of Failed on
				// the grounds that the error reaches the logs; the objection is that Fresh
				// is an affirmative claim of health, not merely an omission from Failed.
				Expect(report.Updated()).To(HaveLen(1))
				Expect(report.Failed()).To(HaveLen(1))
				Expect(report.Fresh()).To(BeEmpty())
			})
		})
	})

	When("dockwatch has been instructed to monitor only", func() {
		When("certain containers are set to monitor only", func() {
			It("should not update those containers", func() {
				client := CreateMockClient(
					&TestData{
						NameOfContainerToKeep: "test-container-02",
						Containers: []types.Container{
							CreateMockContainer(
								"test-container-01",
								"test-container-01",
								"fake-image1:latest",
								time.Now()),
							CreateMockContainerWithConfig(
								"test-container-02",
								"test-container-02",
								"fake-image2:latest",
								false,
								false,
								time.Now(),
								&dockerContainer.Config{
									Labels: map[string]string{
										"io.github.fugginold.dockwatch.monitor-only": "true",
									},
								}),
						},
					},
					false,
					false,
				)
				_, err := actions.Update(client, types.UpdateParams{Cleanup: true})
				Expect(err).NotTo(HaveOccurred())
				Expect(client.TestData.TriedToRemoveImageCount).To(Equal(1))
			})
		})

		When("monitor only is set globally", func() {
			It("should not update any containers", func() {
				client := CreateMockClient(
					&TestData{
						Containers: []types.Container{
							CreateMockContainer(
								"test-container-01",
								"test-container-01",
								"fake-image:latest",
								time.Now()),
							CreateMockContainer(
								"test-container-02",
								"test-container-02",
								"fake-image:latest",
								time.Now()),
						},
					},
					false,
					false,
				)
				_, err := actions.Update(client, types.UpdateParams{Cleanup: true, MonitorOnly: true})
				Expect(err).NotTo(HaveOccurred())
				Expect(client.TestData.TriedToRemoveImageCount).To(Equal(0))
			})
			When("dockwatch has been instructed to have label take precedence", func() {
				It("it should update containers when monitor only is set to false", func() {
					client := CreateMockClient(
						&TestData{
							//NameOfContainerToKeep: "test-container-02",
							Containers: []types.Container{
								CreateMockContainerWithConfig(
									"test-container-02",
									"test-container-02",
									"fake-image2:latest",
									false,
									false,
									time.Now(),
									&dockerContainer.Config{
										Labels: map[string]string{
											"io.github.fugginold.dockwatch.monitor-only": "false",
										},
									}),
							},
						},
						false,
						false,
					)
					_, err := actions.Update(client, types.UpdateParams{Cleanup: true, MonitorOnly: true, LabelPrecedence: true})
					Expect(err).NotTo(HaveOccurred())
					Expect(client.TestData.TriedToRemoveImageCount).To(Equal(1))
				})
				It("it should update not containers when monitor only is set to true", func() {
					client := CreateMockClient(
						&TestData{
							//NameOfContainerToKeep: "test-container-02",
							Containers: []types.Container{
								CreateMockContainerWithConfig(
									"test-container-02",
									"test-container-02",
									"fake-image2:latest",
									false,
									false,
									time.Now(),
									&dockerContainer.Config{
										Labels: map[string]string{
											"io.github.fugginold.dockwatch.monitor-only": "true",
										},
									}),
							},
						},
						false,
						false,
					)
					_, err := actions.Update(client, types.UpdateParams{Cleanup: true, MonitorOnly: true, LabelPrecedence: true})
					Expect(err).NotTo(HaveOccurred())
					Expect(client.TestData.TriedToRemoveImageCount).To(Equal(0))
				})
				It("it should update not containers when monitor only is not set", func() {
					client := CreateMockClient(
						&TestData{
							Containers: []types.Container{
								CreateMockContainer(
									"test-container-01",
									"test-container-01",
									"fake-image:latest",
									time.Now()),
							},
						},
						false,
						false,
					)
					_, err := actions.Update(client, types.UpdateParams{Cleanup: true, MonitorOnly: true, LabelPrecedence: true})
					Expect(err).NotTo(HaveOccurred())
					Expect(client.TestData.TriedToRemoveImageCount).To(Equal(0))
				})

			})
		})
	})

	When("dockwatch has been instructed to run lifecycle hooks", func() {

		When("pre-update script returns 1", func() {
			It("should not update those containers", func() {
				client := CreateMockClient(
					&TestData{
						//NameOfContainerToKeep: "test-container-02",
						Containers: []types.Container{
							CreateMockContainerWithConfig(
								"test-container-02",
								"test-container-02",
								"fake-image2:latest",
								true,
								false,
								time.Now(),
								&dockerContainer.Config{
									Labels: map[string]string{
										"io.github.fugginold.dockwatch.lifecycle.pre-update-timeout": "190",
										"io.github.fugginold.dockwatch.lifecycle.pre-update":         "/PreUpdateReturn1.sh",
									},
									ExposedPorts: map[nat.Port]struct{}{},
								}),
						},
					},
					false,
					false,
				)

				_, err := actions.Update(client, types.UpdateParams{Cleanup: true, LifecycleHooks: true})
				Expect(err).NotTo(HaveOccurred())
				Expect(client.TestData.TriedToRemoveImageCount).To(Equal(0))
			})

		})

		When("prupddate script returns 75", func() {
			It("should not update those containers", func() {
				client := CreateMockClient(
					&TestData{
						//NameOfContainerToKeep: "test-container-02",
						Containers: []types.Container{
							CreateMockContainerWithConfig(
								"test-container-02",
								"test-container-02",
								"fake-image2:latest",
								true,
								false,
								time.Now(),
								&dockerContainer.Config{
									Labels: map[string]string{
										"io.github.fugginold.dockwatch.lifecycle.pre-update-timeout": "190",
										"io.github.fugginold.dockwatch.lifecycle.pre-update":         "/PreUpdateReturn75.sh",
									},
									ExposedPorts: map[nat.Port]struct{}{},
								}),
						},
					},
					false,
					false,
				)
				_, err := actions.Update(client, types.UpdateParams{Cleanup: true, LifecycleHooks: true})
				Expect(err).NotTo(HaveOccurred())
				Expect(client.TestData.TriedToRemoveImageCount).To(Equal(0))
			})

		})

		When("prupddate script returns 0", func() {
			It("should update those containers", func() {
				client := CreateMockClient(
					&TestData{
						//NameOfContainerToKeep: "test-container-02",
						Containers: []types.Container{
							CreateMockContainerWithConfig(
								"test-container-02",
								"test-container-02",
								"fake-image2:latest",
								true,
								false,
								time.Now(),
								&dockerContainer.Config{
									Labels: map[string]string{
										"io.github.fugginold.dockwatch.lifecycle.pre-update-timeout": "190",
										"io.github.fugginold.dockwatch.lifecycle.pre-update":         "/PreUpdateReturn0.sh",
									},
									ExposedPorts: map[nat.Port]struct{}{},
								}),
						},
					},
					false,
					false,
				)
				_, err := actions.Update(client, types.UpdateParams{Cleanup: true, LifecycleHooks: true})
				Expect(err).NotTo(HaveOccurred())
				Expect(client.TestData.TriedToRemoveImageCount).To(Equal(1))
			})
		})

		When("container is linked to restarting containers", func() {
			It("should be marked for restart", func() {

				provider := CreateMockContainerWithConfig(
					"test-container-provider",
					"/test-container-provider",
					"fake-image2:latest",
					true,
					false,
					time.Now(),
					&dockerContainer.Config{
						Labels:       map[string]string{},
						ExposedPorts: map[nat.Port]struct{}{},
					})

				provider.SetStale(true)

				consumer := CreateMockContainerWithConfig(
					"test-container-consumer",
					"/test-container-consumer",
					"fake-image3:latest",
					true,
					false,
					time.Now(),
					&dockerContainer.Config{
						Labels: map[string]string{
							"io.github.fugginold.dockwatch.depends-on": "test-container-provider",
						},
						ExposedPorts: map[nat.Port]struct{}{},
					})

				containers := []types.Container{
					provider,
					consumer,
				}

				Expect(provider.ToRestart()).To(BeTrue())
				Expect(consumer.ToRestart()).To(BeFalse())

				actions.UpdateImplicitRestart(containers)

				Expect(containers[0].ToRestart()).To(BeTrue())
				Expect(containers[1].ToRestart()).To(BeTrue())

			})

		})

		When("container is not running", func() {
			It("skip running preupdate", func() {
				client := CreateMockClient(
					&TestData{
						//NameOfContainerToKeep: "test-container-02",
						Containers: []types.Container{
							CreateMockContainerWithConfig(
								"test-container-02",
								"test-container-02",
								"fake-image2:latest",
								false,
								false,
								time.Now(),
								&dockerContainer.Config{
									Labels: map[string]string{
										"io.github.fugginold.dockwatch.lifecycle.pre-update-timeout": "190",
										"io.github.fugginold.dockwatch.lifecycle.pre-update":         "/PreUpdateReturn1.sh",
									},
									ExposedPorts: map[nat.Port]struct{}{},
								}),
						},
					},
					false,
					false,
				)
				_, err := actions.Update(client, types.UpdateParams{Cleanup: true, LifecycleHooks: true})
				Expect(err).NotTo(HaveOccurred())
				Expect(client.TestData.TriedToRemoveImageCount).To(Equal(1))
			})

		})

		When("container is restarting", func() {
			It("skip running preupdate", func() {
				client := CreateMockClient(
					&TestData{
						//NameOfContainerToKeep: "test-container-02",
						Containers: []types.Container{
							CreateMockContainerWithConfig(
								"test-container-02",
								"test-container-02",
								"fake-image2:latest",
								false,
								true,
								time.Now(),
								&dockerContainer.Config{
									Labels: map[string]string{
										"io.github.fugginold.dockwatch.lifecycle.pre-update-timeout": "190",
										"io.github.fugginold.dockwatch.lifecycle.pre-update":         "/PreUpdateReturn1.sh",
									},
									ExposedPorts: map[nat.Port]struct{}{},
								}),
						},
					},
					false,
					false,
				)
				_, err := actions.Update(client, types.UpdateParams{Cleanup: true, LifecycleHooks: true})
				Expect(err).NotTo(HaveOccurred())
				Expect(client.TestData.TriedToRemoveImageCount).To(Equal(1))
			})

		})

	})
})

// A removal the daemon accepted but had not finished within the budget used to be
// recorded as a failed stop, so the recreate was skipped -- and the daemon finished
// the removal a moment later regardless, leaving nothing behind. Attempting the
// recreate is safe: docker enforces name uniqueness, so if the container really is
// still there the create fails and the original is left untouched.
var _ = Describe("a container whose removal was not confirmed", func() {
	It("should still be recreated", func() {
		testData := getCommonTestData("")
		testData.StopErrors = map[string]error{
			"test-container-01": container.ErrRemovalUnconfirmed,
		}
		client := CreateMockClient(testData, false, false)

		_, err := actions.Update(client, types.UpdateParams{})
		Expect(err).NotTo(HaveOccurred())

		Expect(testData.StartedContainers).To(ContainElement("test-container-01"),
			"an unconfirmed removal must not skip the recreate; the container is otherwise lost")
	})

	// The rolling-restart path has its own stop/restart loop, so a fix applied only
	// to the batched path leaves this one still losing containers.
	It("should still be recreated during a rolling restart", func() {
		testData := getCommonTestData("")
		testData.StopErrors = map[string]error{
			"test-container-01": container.ErrRemovalUnconfirmed,
		}
		client := CreateMockClient(testData, false, false)

		_, err := actions.Update(client, types.UpdateParams{RollingRestart: true})
		Expect(err).NotTo(HaveOccurred())

		Expect(testData.StartedContainers).To(ContainElement("test-container-01"),
			"the rolling-restart path must handle an unconfirmed removal too")
	})
})
